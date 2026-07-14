package users

// import (
//     database "github.com/dev-karani/chirpy/internal/database"
// )

// type Service struct {
//     db        *database.Queries
//     jwtSecret string
// }

// func NewService(db *database.Queries, jwtSecret string) *Service {
//     return &Service{
//         db: db,
//         jwtSecret: jwtSecret,
//     }
// }


// type LoginResult struct {
// 	User
// 	Token        string
// 	RefreshToken string
// }

// // func (s *Service) Login(ctx context.Context, email string,password string) (User, error) {
// // 	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
// // 	if err != nil {
// // 		httpAPI.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
// // 		return
// // 	}
// // 	match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
// // 	if err != nil || !match {
// // 		httpAPI.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
// // 		return
// // 	}
	
// // 	jwtToken, err := auth.MakeJWT(user.ID, h.jwtSecret, time.Hour)
// // 	if err != nil {
// // 		httpAPI.RespondWithError(w, http.StatusUnauthorized, "failed to generate jwt")
// // 		return
// // 	}

// // 	//refreshToken
// // 	refreshToken, err := auth.MakeRefreshToken()
// // 	if err != nil {
// // 		httpAPI.RespondWithError(w, http.StatusInternalServerError, "couldnt create refresh token")
// // 	}

// // 	now := time.Now().UTC()

// // 	_, err = h.db.CreateRefreshToken(
// // 		r.Context(),
// // 		database.CreateRefreshTokenParams{
// // 			Token:     refreshToken,
// // 			CreatedAt: now,
// // 			UpdatedAt: now,
// // 			UserID:    user.ID,
// // 			ExpiresAt: now.Add(60 * 24 * time.Hour),
// // 		},
// // 	)

// // }


// // func (s *Service) CreateUser(ctx context.Context, email string,password string) (User, error) {


// // }

// func (s *Service) CreateUser(
// 	ctx context.Context,
// 	email string,
// 	password string,
// ) (User, error) {

// 	hashedPassword, err := auth.HashPassword(password)
// 	if err != nil {
// 		return User{}, err
// 	}

// 	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
// 		Email:          email,
// 		HashedPassword: hashedPassword,
// 	})
// 	if err != nil {
// 		return User{}, err
// 	}

// 	return User{
// 		ID:          user.ID,
// 		CreatedAt:   user.CreatedAt,
// 		UpdatedAt:   user.UpdatedAt,
// 		Email:       user.Email,
// 		IsChirpyRed: user.IsChirpyRed,
// 	}, nil
// }


// func (s *Service) Login(
// 	ctx context.Context,
// 	email string,
// 	password string,
// ) (LoginResult, error) {

// 	user, err := s.db.GetUserByEmail(ctx, email)
// 	if err != nil {
// 		return LoginResult{}, err
// 	}

// 	match, err := auth.CheckPasswordHash(password, user.HashedPassword)
// 	if err != nil || !match {
// 		return LoginResult{}, err
// }

// 	jwtToken, err := auth.MakeJWT(
// 		user.ID,
// 		s.jwtSecret,
// 		time.Hour,
// 	)
// 	if err != nil {
// 		return LoginResult{}, err
// 	}

// 	refreshToken, err := auth.MakeRefreshToken()
// 	if err != nil {
// 		return LoginResult{}, err
// 	}

// 	now := time.Now().UTC()

// 	_, err = s.db.CreateRefreshToken(
// 		ctx,
// 		database.CreateRefreshTokenParams{
// 			Token:     refreshToken,
// 			CreatedAt: now,
// 			UpdatedAt: now,
// 			UserID:    user.ID,
// 			ExpiresAt: now.Add(60 * 24 * time.Hour),
// 		},
// 	)
// 	if err != nil {
// 		return LoginResult{}, err
// 	}

// 	return LoginResult{
// 		User: User{
// 			ID:          user.ID,
// 			CreatedAt:   user.CreatedAt,
// 			UpdatedAt:   user.UpdatedAt,
// 			Email:       user.Email,
// 			IsChirpyRed: user.IsChirpyRed,
// 		},
// 		Token:        jwtToken,
// 		RefreshToken: refreshToken,
// 	}, nil
// }