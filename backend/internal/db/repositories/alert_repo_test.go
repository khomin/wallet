package repositories

import (
	"testing"
	"time"
	"tracker/internal/core/domain"
)

func TestAlertRepo(t *testing.T) {
	ctx, db, err := prepare()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewAlertRepository(db)
	userRepo := NewUserRepo(db)

	expectedUser := domain.User{
		ID:    "demo",
		Name:  "Demo",
		Email: "demo@demo.com",
	}
	expectedSymbol := "ETH"

	//
	// new user first
	//
	err = userRepo.EnsureExists(ctx, expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	//
	// create
	//
	resultAlert, err := repo.Create(ctx, domain.Alert{
		UserID:     expectedUser.ID,
		CoinSymbol: expectedSymbol,
		Condition:  domain.AlertConditionAbove,
		Enabled:    true,
		Price:      2300,
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resultAlert.ID == "" || resultAlert.UserID != expectedUser.ID || resultAlert.CoinSymbol != expectedSymbol {
		t.Fatal("expected value")
	}

	//
	// update
	//
	resultAlertUpdate, err := repo.Update(ctx, expectedUser.ID, resultAlert.ID, domain.AlertUpdate{
		Condition: domain.AlertConditionBelow,
		Price:     2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resultAlertUpdate.ID == "" || resultAlertUpdate.UserID != expectedUser.ID || resultAlertUpdate.CoinSymbol != expectedSymbol {
		t.Fatal("expected value")
	}

	//
	// disable
	//
	resultAlertDisable, err := repo.DisableAsCompleted(ctx, expectedUser.ID, resultAlert.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resultAlertDisable.ID == "" || resultAlertDisable.UserID != expectedUser.ID || resultAlertDisable.CoinSymbol != expectedSymbol {
		t.Fatal("expected value")
	}

	//
	// pause
	//
	resultAlertPause, err := repo.Pause(ctx, expectedUser.ID, resultAlert.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resultAlertPause.ID == "" || resultAlertPause.UserID != expectedUser.ID || resultAlertPause.CoinSymbol != expectedSymbol {
		t.Fatal("expected value")
	}

	//
	// enable
	//
	resultAlertEnable, err := repo.Enable(ctx, expectedUser.ID, resultAlert.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resultAlertEnable.ID == "" || resultAlertEnable.UserID != expectedUser.ID || resultAlertEnable.CoinSymbol != expectedSymbol {
		t.Fatal("expected value")
	}

	//
	// list by user
	//
	resultAlertList, err := repo.ListByUser(ctx, expectedUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(resultAlertList) == 0 {
		t.Fatal("expected value")
	}

	//
	// list active
	//
	resultAlertListActive, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(resultAlertListActive) == 0 {
		t.Fatal("expected value")
	}

	//
	// delete
	//
	err = repo.Delete(ctx, expectedUser.ID, resultAlert.ID)
	if err != nil {
		t.Fatal(err)
	}
}
