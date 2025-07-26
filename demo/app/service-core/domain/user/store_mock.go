package user

type MockStore struct{}

func NewMockStore() *MockStore {
	return &MockStore{}
}

// const mockAccess = 0x0000000000000001

// func (s *MockStore) SelectTokenByID(_ context.Context, id string) (*query.Token, error) {
// 	switch id {
// 	case "1":
// 		return &query.Token{
// 			ID:        "1",
// 			Expires:   "2050-01-01T00:00:00Z",
// 			UserID:    "1",
// 			State:     "",
// 			Verifier:  "",
// 			ReturnURL: "",
// 		}, nil
// 	case "2":
// 		return &query.Token{
// 			ID:        "2",
// 			Expires:   "2000-01-01T00:00:00Z",
// 			UserID:    "1",
// 			State:     "",
// 			Verifier:  "",
// 			ReturnURL: "",
// 		}, nil
// 	case "3":
// 		return &query.Token{
// 			ID:        "3",
// 			Expires:   "2050-01-01T00:00:00Z",
// 			UserID:    "",
// 			State:     "",
// 			Verifier:  "",
// 			ReturnURL: "",
// 		}, nil
// 	}
// 	return nil, errors.New("invalid token by ID")
// }
// 
// func (s *MockStore) SelectTokenByState(_ context.Context, _ string) (*Token, error) {
// 	return &query.Token{
// 		ID:        "1",
// 		Expires:   "2050-01-01T00:00:00Z",
// 		UserID:    "1",
// 		State:     "1",
// 		Verifier:  "1",
// 		ReturnURL: "1",
// 	}, nil
// }
// 
// func (s *MockStore) InsertToken(_ context.Context, _ string, _ string, _ string, _ string, _ string) (*query.Token, error) {
// 	return &query.Token{
// 		ID:        "1",
// 		Expires:   "2050-01-01T00:00:00Z",
// 		UserID:    "1",
// 		State:     "",
// 		Verifier:  "",
// 		ReturnURL: "",
// 	}, nil
// }
// 
// func (s *MockStore) UpdateTokenExpires(_ context.Context, _ string, _ string) error {
// 	return nil
// }
// 
// func (s *MockStore) SelectUserByID(_ context.Context, id string) (*query.User, error) {
//     if id == "1" {
// 		return &query.User{
// 			ID:              "1",
// 			Created:         "",
// 			Updated:         "",
// 			Email:           "admin@gofast.live",
// 			Access:          mockAccess,
// 			Sub:             "",
// 			Avatar:          "",
// 			SubscriptionID:  "",
// 			SubscriptionEnd: "",
// 		}, nil
// 	}
// 	return nil, errors.New("missing user")
// }
// 
// func (s *MockStore) SelectUserByEmailAndSub(_ context.Context, _ string, _ string) (*query.User, error) {
// 	return &query.User{
// 		ID:              "1",
// 		Created:         "",
// 		Updated:         "",
// 		Email:           "",
// 		Access:          mockAccess,
// 		Sub:             "",
// 		Avatar:          "",
// 		SubscriptionID:  "",
// 		SubscriptionEnd: "",
// 	}, nil
// }
// 
// func (s *MockStore) InsertUser(_ context.Context, _ string, _ int64, _ string, _ string) (*query.User, error) {
// 	return &query.User{
// 		ID:              "1",
// 		Created:         "",
// 		Updated:         "",
// 		Email:           "",
// 		Access:          mockAccess,
// 		Sub:             "",
// 		Avatar:          "",
// 		SubscriptionID:  "",
// 		SubscriptionEnd: "",
// 	}, nil
// }
// 
// func (s *MockStore) UpdateUserActivity(_ context.Context, _ string) error {
// 	return nil
// }
