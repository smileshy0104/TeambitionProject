package jwts

import "testing"

func TestParseToken(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQwMjczNzEsImlwIjoiMTI3LjAuMC4xIiwidG9rZW4iOiIxMDAxIn0.gPwT0v07YhmbHfhX0MPyT5Bmtd39ahzeVFJQ9rqbhb8"
	ParseToken(tokenString, "msproject", "127.0.0.1")
}
