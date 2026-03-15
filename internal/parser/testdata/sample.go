package sample

import (
	"fmt"
	"net/http"
)

// User represents a user in the system.
type User struct {
	Name  string
	Email string
	Age   int
}

// Stringer is an interface for types that can convert to string.
type Stringer interface {
	String() string
}

// HandleRequest processes an incoming HTTP request
// and returns an appropriate response.
func HandleRequest(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "hello")
	return nil
}

// String returns the user's name.
func (u *User) String() string {
	return u.Name
}

// Greet returns a greeting for the user.
func (u *User) Greet(greeting string) string {
	return fmt.Sprintf("%s, %s!", greeting, u.Name)
}

func helperFunc() {
	// unexported helper
}

// TestHandleRequest tests the HandleRequest function.
func TestHandleRequest(t interface{ Fatal(...interface{}) }) {
	// test body
}
