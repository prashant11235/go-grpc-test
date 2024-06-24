package main

import (
	"context"
	"testing"

	pb "test.com/traintix/pb"
)

func TestPurchaseTicket_Success(t *testing.T) {
	server := NewServer()

	ctx := context.Background()
	user := &pb.User{
		FirstName: "Temp",
		LastName:  "Temp",
		Email:     "temp.temp@temp.com",
	}
	purchaseRequest := &pb.PurchaseRequest{User: user}

	receipt, err := server.PurchaseTicket(ctx, purchaseRequest)
	if err != nil {
		t.Fatalf("PurchaseTicket() error = %v", err)
	}

	if receipt.User.Email != user.Email {
		t.Errorf("Expected user email %v, got %v", user.Email, receipt.User.Email)
	}
	if receipt.SectionId != 1 {
		t.Errorf("Expected SectionId 1, got %d", receipt.SectionId)
	}
	if receipt.SeatId != 0 {
		t.Errorf("Expected SeatId 0, got %d", receipt.SeatId)
	}
	if _, ok := server.emptySeats[0][0]; ok {
		t.Errorf("Seat 0 in Section 0 should be occupied")
	}
}

func TestGetReceiptDetails(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	user := &pb.User{
		FirstName: "Temp",
		LastName:  "Temp",
		Email:     "temp.temp@temp.com",
	}

	receipt := &pb.Receipt{
		User:      user,
		SectionId: 1,
	}
	server.userReceipts[user.Email] = receipt

	retrievedReceipt, err := server.GetReceiptDetails(ctx, user)
	if err != nil {
		t.Fatalf("GetReceiptDetails() error = %v", err)
	}

	if retrievedReceipt.SectionId != receipt.SectionId {
		t.Errorf("GetReceiptDetails() = %v; want %v", retrievedReceipt, receipt)
	}
}

func TestRemoveUser(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	user := &pb.User{
		FirstName: "Temp",
		LastName:  "Temp",
		Email:     "temp.temp@temp.com",
	}
	receipt := &pb.Receipt{
		User:      user,
		SectionId: 1,
	}
	server.userReceipts[user.Email] = receipt
	server.sectionAssignments[0][1] = user
	server.emptySeats[0][1] = false

	operationResult, err := server.RemoveUser(ctx, user)
	if err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
	if !operationResult.Result {
		t.Errorf("Expected operation result to be true")
	}
	if _, ok := server.userReceipts[user.Email]; ok {
		t.Errorf("Expected user receipt to be removed")
	}
}

func TestModifyUserSeat(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	user := &pb.User{
		FirstName: "Temp",
		LastName:  "Temp",
		Email:     "temp.temp@temp.com",
	}
	receipt := &pb.Receipt{
		User:      user,
		SectionId: 1,
	}
	server.userReceipts[user.Email] = receipt

	modifyUserSeatRequest := &pb.ModifyUserSeatRequest{
		User: user,
		Seat: &pb.Seat{SeatId: 2},
	}

	newReceipt, err := server.ModifyUserSeat(ctx, modifyUserSeatRequest)
	if err != nil {
		t.Fatalf("ModifyUserSeat() error = %v", err)
	}

	if newReceipt.SeatId != 2 {
		t.Errorf("Expected new SeatId 2, got %d", newReceipt.SeatId)
	}
	if server.sectionAssignments[1][2].Email != user.Email {
		t.Errorf("Expected seat 2 to be assigned to %v, got %v", user.Email, server.sectionAssignments[0][2].Email)
	}
}
