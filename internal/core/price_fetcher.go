package core

import (
	"context"
	"time"
	"tracker/internal/cache"
	"tracker/internal/client"
	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type PriceFetcher struct {
	coingeckoClient *client.CoinGeckoClient
	alchemyClient   *client.PriceClient
	cache           *cache.RedisClient
	repo            *repositories.PriceRepository
	// repo            *repositories.CoinSnapshotRepository
	interval time.Duration
}

func NewPriceFetcher(
	coingeckoClient *client.CoinGeckoClient,
	alchemyClient *client.PriceClient,
	cache *cache.RedisClient,
	repo *repositories.PriceRepository,
	interval time.Duration,
) *PriceFetcher {
	return &PriceFetcher{
		coingeckoClient: coingeckoClient,
		alchemyClient:   alchemyClient,
		cache:           cache,
		repo:            repo,
		interval:        interval,
	}
}

// Start begins the background fetching loop
func (f *PriceFetcher) Start(ctx context.Context) {
	logrus.Info("Starting price fetcher background worker...")

	// Run immediately on startup
	f.fetchAndStore(ctx)

	// Then run every interval
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Price fetcher stopped")
			return
		case <-ticker.C:
			f.fetchAndStore(ctx)
		}
	}
}

func (f *PriceFetcher) fetchAndStore(ctx context.Context) {
	logrus.Info("Fetching latest prices...")
	startTime := time.Now()

	// 1. Fetch all 250 coins from CoinGecko
	coins, err := f.coingeckoClient.GetCoins(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch coins from CoinGecko")
		return
	}

	// 2. Extract symbols for Alchemy
	var symbols []string
	for _, coin := range coins {
		symbols = append(symbols, coin.Symbol)
	}

	// 3. Fetch real-time prices from Alchemy (batch in chunks of 25)
	pricesMap, err := f.fetchPricesFromAlchemy(ctx, symbols)
	if err != nil {
		logrus.WithError(err).Warn("Failed to fetch prices from Alchemy, using CoinGecko prices only")
	}

	// 4. Enrich coins with Alchemy prices
	enrichedCoins := f.enrichWithAlchemyPrices(coins, pricesMap)

	// 5. Store in Redis cache
	if err := f.cache.SetJSON(ctx, "prices:all", enrichedCoins, 60); err != nil {
		logrus.WithError(err).Error("Failed to cache prices in Redis")
	} else {
		logrus.WithField("count", len(enrichedCoins)).Debug("Stored prices in Redis cache")
	}

	// 6. Store in Postgres (async, don't block)
	go f.storeSnapshots(context.Background(), enrichedCoins)

	elapsed := time.Since(startTime)
	logrus.WithField("elapsed_ms", elapsed.Milliseconds()).Info("Price fetch completed")
}

func (f *PriceFetcher) fetchPricesFromAlchemy(ctx context.Context, symbols []string) (map[string]float64, error) {
	pricesMap := make(map[string]float64)
	// Alchemy has a limit of 25 symbols per request
	chunks := lo.Chunk(symbols, 25)
	for i := 0; i < len(chunks); i++ {
		chunk := chunks[i]
		prices, err := f.alchemyClient.GetPrices(ctx, chunk)
		if err != nil {
			logrus.WithError(err).WithField("chunk", chunk).Warn("Failed to fetch price chunk")
			continue
		}
		for _, p := range prices {
			pricesMap[p.Symbol] = p.PriceUSD
		}
	}
	return pricesMap, nil
}

func (f *PriceFetcher) enrichWithAlchemyPrices(coins []client.CoinGeckoCoin, pricesMap map[string]float64) []client.CoinGeckoCoin {
	enriched := make([]client.CoinGeckoCoin, len(coins))

	for i, coin := range coins {
		// If Alchemy has a price, use it (more accurate/real-time)
		if alchemyPrice, ok := pricesMap[coin.Symbol]; ok {
			coin.CurrentPrice = alchemyPrice
		}
		// Otherwise keep CoinGecko price
		enriched[i] = coin
	}

	return enriched
}

func (f *PriceFetcher) storeSnapshots(ctx context.Context, coins []client.CoinGeckoCoin) {
	logrus.WithField("count", len(coins)).Debug("Storing snapshots in Postgres...")

	// Batch insert for performance
	var snapshots []models.CoinSnapshot
	for _, coin := range coins {
		snapshot := models.CoinSnapshot{
			ID:                        uuid.New(),
			CoinID:                    coin.ID,
			Symbol:                    coin.Symbol,
			Name:                      coin.Name,
			PriceUSD:                  coin.CurrentPrice,
			MarketCapUSD:              coin.MarketCap,
			MarketCapRank:             coin.MarketCapRank,
			TotalVolumeUSD:            coin.TotalVolume,
			PriceChange24h:            coin.PriceChange24h,
			PriceChangePercent24h:     coin.PriceChangePercent24h,
			MarketCapChange24h:        coin.MarketCapChange24h,
			MarketCapChangePercent24h: coin.MarketCapChangePercent24h,
			CirculatingSupply:         coin.CirculatingSupply,
			TotalSupply:               coin.TotalSupply,
			MaxSupply:                 coin.MaxSupply,
			ATH:                       coin.ATH,
			ATHChangePercent:          coin.ATHChangePercent,
			ATHDate:                   coin.ATHDate,
			ATL:                       coin.ATL,
			ATLChangePercent:          coin.ATLChangePercent,
			ATLDate:                   coin.ATLDate,
			ImageURL:                  coin.Image,
			LastUpdated:               coin.LastUpdated,
			SnapshotAt:                time.Now(),
		}
		snapshots = append(snapshots, snapshot)
	}

	if err := f.repo.SetCoinSnapshot(ctx, snapshots); err != nil {
		logrus.WithError(err).Error("Failed to store snapshots in Postgres")
	} else {
		logrus.WithField("count", len(snapshots)).Debug("Stored snapshots in Postgres")
	}
}
