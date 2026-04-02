package state

type AuthStateRepository struct {
	path string
}

func NewAuthStateRepository(path string) *AuthStateRepository {
	return &AuthStateRepository{path: path}
}

func (r *AuthStateRepository) Get() (AuthState, error) {
	return ReadAuthState(r.path)
}

func (r *AuthStateRepository) Save(st AuthState) error {
	return WriteAuthState(r.path, st)
}
