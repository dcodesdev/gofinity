package main

import "fmt"

// Celsius is a temperature that travels as a JSON *string* with a unit
// suffix - 21.5 encodes as "21.5C" - rather than as a bare number.
//
// A type controls its own JSON by implementing these two methods. Once it
// does, every struct that holds one gets the behaviour for free: encoding/json
// checks for the interfaces before it falls back to reflection.
type Celsius float64

// MarshalJSON returns the JSON encoding of c: the number, formatted with no
// trailing zeroes, followed by a C, as a JSON string.
//
//	Celsius(21.5)  ->  "21.5C"
//	Celsius(0)     ->  "0C"
//	Celsius(-3.25) ->  "-3.25C"
//
// The return value is a complete JSON value, quotes included - not the text
// inside the quotes. A value receiver, so both a Celsius and a *Celsius have
// the method.
func (c Celsius) MarshalJSON() ([]byte, error) {
	// TODO
	return nil, nil
}

// UnmarshalJSON parses the form MarshalJSON produces back into c.
//
// The receiver is a **pointer**, because there is nowhere else to put the
// result. That asymmetry is the one thing to remember about this pair.
//
// A value that is not a JSON string, or a string without the C suffix, or one
// whose number does not parse, is an error.
func (c *Celsius) UnmarshalJSON(data []byte) error {
	// TODO
	return nil
}

// Address is nested inside a Profile, and its keys are lowercase too. Nesting
// needs no special handling: the encoder recurses, and the tags on this struct
// apply wherever it appears.
type Address struct {
	City    string
	Zip     string
	Country string
}

// Profile is the whole exercise. Every field below needs a tag, and the tests
// pin the exact bytes, so read the required shape carefully.
//
// A full profile encodes as, with the keys in this order:
//
//	{"username":"ada","email":"ada@example.com","age":36,"nickname":"count",
//	 "address":{"city":"London","zip":"E1","country":"UK"},
//	 "roles":["admin"],"temp":"21.5C"}
//
// An empty one encodes as:
//
//	{"username":"","address":{"city":"","country":""},"temp":"0C"}
//
// Which is to say:
//
//   - username, address and temp are always present, even when empty.
//   - email, age, nickname and roles disappear when they are empty.
//   - zip disappears when it is empty, city and country do not.
//   - password never appears at all, in either direction - it is a secret that
//     happens to live in the same struct.
//   - notes is unexported, so it was never going anywhere.
type Profile struct {
	Username string
	Email    string
	Password string
	Age      int
	// Nickname is a *string rather than a string so that "absent" and "set to
	// the empty string" stay different. With omitempty, a nil pointer is
	// dropped and a pointer to "" is written as "".
	Nickname *string
	Address  Address
	Roles    []string
	Temp     Celsius

	notes string
}

// Encode returns the JSON encoding of p, unindented.
func Encode(p Profile) (string, error) {
	// TODO
	return "", nil
}

// Decode parses a profile.
func Decode(data []byte) (Profile, error) {
	// TODO
	return Profile{}, nil
}

// RoundTrip encodes p and decodes the result straight back.
//
// What comes out is not always what went in, and that is the point: anything
// tagged "-" or left unexported is dropped on the way through.
func RoundTrip(p Profile) (Profile, error) {
	// TODO
	return Profile{}, nil
}

func main() {
	out, err := Encode(Profile{Username: "ada", Temp: 21.5})
	if err != nil {
		fmt.Println("encode failed:", err)
		return
	}
	fmt.Println(out)
}
