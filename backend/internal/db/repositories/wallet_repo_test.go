package repositories

import (
	"testing"

	"github.com/google/uuid"
)

func TestWalletRepo(t *testing.T) {
	ctx, db, err := Prepare()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewWalletRepository(db)
	userRepo := NewUserRepo(db)

	expectedUser := DemoUser()

	//
	// new user first
	//
	err = userRepo.EnsureExists(ctx, &expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	//
	// delete from previous tests
	//
	resultList, err := repo.List(ctx, expectedUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range resultList {
		id, _ := uuid.Parse(i.ID)
		repo.Delete(ctx, expectedUser.ID, id)
	}

	//
	// create
	//
	resultCreate, err := repo.Create(ctx, expectedUser.ID, "ETH", "0xEC2dFb47E5AA06da508D816D83b4833f6eBE9532", "ETH", "Test")
	if err != nil {
		t.Fatal(err)
	}
	if resultCreate.ID == "" || resultCreate.UserID != expectedUser.ID || resultCreate.Symbol != "ETH" || resultCreate.Address != "0xEC2dFb47E5AA06da508D816D83b4833f6eBE9532" {
		t.Fatal("expected correct value")
	}

	id, _ := uuid.Parse(resultCreate.ID)

	//
	// update
	//
	resultUpdate, err := repo.Update(ctx, expectedUser.ID, id, "Test updated")
	if err != nil {
		t.Fatal(err)
	}
	if resultUpdate.ID == "" || resultUpdate.UserID != expectedUser.ID || resultUpdate.Symbol != "ETH" || resultUpdate.Address != "0xEC2dFb47E5AA06da508D816D83b4833f6eBE9532" || resultUpdate.Label != "Test updated" {
		t.Fatal("expected correct value")
	}

	//
	// update balance
	//
	err = repo.UpdateBalance(ctx, expectedUser.ID, id, 123, 456)
	if err != nil {
		t.Fatal(err)
	}

	//
	// get
	//
	resultGet, err := repo.Get(ctx, expectedUser.ID, id)
	if err != nil {
		t.Fatal(err)
	}
	if resultGet.ID == "" || resultGet.UserID != expectedUser.ID || resultGet.Symbol != "ETH" || resultGet.Balance != 123 || resultGet.BalanceUSD != 456 {
		t.Fatal("expected correct value")
	}

	//
	// list
	//
	resultList, err = repo.List(ctx, expectedUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(resultList) == 0 {
		t.Fatal("expected >= 1")
	}

	//
	// list for sync
	//
	_, err = repo.ListForSync(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	//
	// delete
	//
	err = repo.Delete(ctx, expectedUser.ID, id)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("completed")
}
