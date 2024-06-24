package main

import (
	"context"
	"sync"

	pb "test.com/traintix/pb"
)

const numSections = 2
const numSeats = 50

type tixMgrServer struct {
	pb.UnimplementedTixMgrServer

	mu sync.Mutex // protects section assignments

	userReceipts        map[string]*pb.Receipt
	sectionAssignments  []map[int32]*pb.User
	lastSectionAssigned int32
	emptySeats          []map[int]bool
}

func (s *tixMgrServer) PurchaseTicket(ctx context.Context, purchaseRequest *pb.PurchaseRequest) (*pb.Receipt, error) {
	user := &pb.User{
		FirstName: purchaseRequest.User.FirstName,
		LastName:  purchaseRequest.User.LastName,
		Email:     purchaseRequest.User.Email,
	}

	// Lock the mutex to ensure correct seat and section assignment
	s.mu.Lock()
	defer s.mu.Unlock()

	// update section assignment
	sectionToAssign := (s.lastSectionAssigned + 1) % numSections // ensure round-robin assignment - alternative could be to greedily fill sections
	s.lastSectionAssigned = sectionToAssign
	seatSeq := int32(s.getEmptySeat(int(sectionToAssign)))

	s.sectionAssignments[sectionToAssign][seatSeq] = user
	delete(s.emptySeats[sectionToAssign], int(seatSeq)) // remove seat from list of empty seats

	// generate receipt
	receipt := &pb.Receipt{
		User:      user,
		SeatId:    seatSeq,
		SectionId: sectionToAssign,
	}

	// update user-receipt map
	s.userReceipts[user.Email] = receipt

	return receipt, nil
}

// GetFeature returns the feature at the given point.
func (s *tixMgrServer) GetReceiptDetails(ctx context.Context, user *pb.User) (*pb.Receipt, error) {
	receipt, ok := s.userReceipts[user.Email]
	if ok {
		return receipt, nil
	}

	// if no receipt find return empty response
	return &pb.Receipt{}, nil
}

// GetSectionDetails returns all assignments for a section
func (s *tixMgrServer) GetSectionDetails(ctx context.Context, section *pb.Section) (*pb.SectionDetails, error) {
	section_map := s.sectionAssignments[section.SectionId]

	sectionDetails := &pb.SectionDetails{
		SeatMap: section_map,
	}
	return sectionDetails, nil
}

func (s *tixMgrServer) RemoveUser(ctx context.Context, user *pb.User) (*pb.OperationResult, error) {
	// Lock the mutex to ensure correct seat and section assignment
	s.mu.Lock()
	defer s.mu.Unlock()

	receipt := s.userReceipts[user.Email]

	// remove from user receipts
	delete(s.userReceipts, user.Email)

	// update section details
	s.sectionAssignments[receipt.SectionId][receipt.SeatId] = nil
	s.emptySeats[receipt.SectionId][int(receipt.SeatId)] = true

	operationResult := &pb.OperationResult{
		Result: true,
	}

	return operationResult, nil
}

func (s *tixMgrServer) ModifyUserSeat(ctx context.Context, modifyUserSeatRequest *pb.ModifyUserSeatRequest) (*pb.Receipt, error) {
	// Lock the mutex to ensure correct seat and section assignment
	s.mu.Lock()
	defer s.mu.Unlock()

	user := modifyUserSeatRequest.User
	receipt := s.userReceipts[user.Email]

	// update section and seat details details
	section_map := s.sectionAssignments[receipt.SectionId]
	section_map[modifyUserSeatRequest.Seat.SeatId] = user
	section_map[receipt.SeatId] = nil // make old seat nil

	s.emptySeats[receipt.SectionId][int(receipt.SeatId)] = true
	delete(s.emptySeats[receipt.SectionId], int(receipt.SeatId))

	// update user receipts
	receipt.SeatId = modifyUserSeatRequest.Seat.SeatId

	return receipt, nil
}

func (s *tixMgrServer) getEmptySeat(sectionId int) int {
	for k := range s.emptySeats[sectionId] {
		return k
	}
	return -1
}

// loadFeatures loads features from a JSON file.
func (s *tixMgrServer) initializeDetails() {
	s.userReceipts = make(map[string]*pb.Receipt, 15)
	s.sectionAssignments = make([]map[int32]*pb.User, numSections)
	s.emptySeats = make([]map[int]bool, numSections)

	for i := 0; i < numSections; i++ {
		s.sectionAssignments[i] = make(map[int32]*pb.User, numSeats)
		s.emptySeats[i] = make(map[int]bool, numSeats)

		for j := 0; j < numSeats; j++ {
			s.sectionAssignments[i][int32(j)] = nil
			s.emptySeats[i][j] = true
		}
	}

	s.lastSectionAssigned = 0
}

func NewServer() *tixMgrServer {
	s := &tixMgrServer{}
	s.initializeDetails()
	return s
}
