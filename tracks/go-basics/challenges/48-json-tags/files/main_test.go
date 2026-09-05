package main

import (
	"strings"
	"testing"
)

func ptr(s string) *string { return &s }

func fullProfile() Profile {
	return Profile{
		Username: "ada",
		Email:    "ada@example.com",
		Password: "hunter2",
		Age:      36,
		Nickname: ptr("count"),
		Address:  Address{City: "London", Zip: "E1", Country: "UK"},
		Roles:    []string{"admin"},
		Temp:     21.5,
		notes:    "private",
	}
}

func TestCelsiusMarshalsAsAString(t *testing.T) {
	cases := []struct {
		in   Celsius
		want string
	}{
		{21.5, `"21.5C"`},
		{0, `"0C"`},
		{-3.25, `"-3.25C"`},
		{100, `"100C"`},
	}
	for _, c := range cases {
		got, err := c.in.MarshalJSON()
		if err != nil {
			t.Fatalf("Celsius(%v).MarshalJSON returned an error: %v", float64(c.in), err)
		}
		if string(got) != c.want {
			t.Errorf("Celsius(%v).MarshalJSON = %s, want %s", float64(c.in), got, c.want)
		}
	}
}

func TestCelsiusUnmarshals(t *testing.T) {
	cases := []struct {
		in   string
		want Celsius
	}{
		{`"21.5C"`, 21.5},
		{`"0C"`, 0},
		{`"-3.25C"`, -3.25},
	}
	for _, c := range cases {
		var got Celsius
		if err := got.UnmarshalJSON([]byte(c.in)); err != nil {
			t.Fatalf("UnmarshalJSON(%s) returned an error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", c.in, float64(got), float64(c.want))
		}
	}
}

func TestCelsiusUnmarshalRejectsBadInput(t *testing.T) {
	bad := []string{`21.5`, `"21.5"`, `"warmC"`, `"C"`, `null`, `"21.5F"`}
	for _, in := range bad {
		var got Celsius
		if err := got.UnmarshalJSON([]byte(in)); err == nil {
			t.Errorf("UnmarshalJSON(%s) returned nil error, want an error", in)
		}
	}
}

func TestEncodeFullProfile(t *testing.T) {
	got, err := Encode(fullProfile())
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	want := `{"username":"ada","email":"ada@example.com","age":36,"nickname":"count",` +
		`"address":{"city":"London","zip":"E1","country":"UK"},` +
		`"roles":["admin"],"temp":"21.5C"}`
	if got != want {
		t.Errorf("Encode =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeEmptyProfile(t *testing.T) {
	got, err := Encode(Profile{})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	want := `{"username":"","address":{"city":"","country":""},"temp":"0C"}`
	if got != want {
		t.Errorf("Encode(zero) =\n  %s\nwant\n  %s", got, want)
	}
}

func TestEncodeNeverWritesThePassword(t *testing.T) {
	got, err := Encode(fullProfile())
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if strings.Contains(got, "hunter2") || strings.Contains(strings.ToLower(got), "password") {
		t.Errorf("Encode leaked the password: %s", got)
	}
}

func TestEncodeNeverWritesTheUnexportedField(t *testing.T) {
	got, err := Encode(fullProfile())
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if strings.Contains(got, "private") || strings.Contains(strings.ToLower(got), "notes") {
		t.Errorf("Encode leaked the unexported field: %s", got)
	}
}

func TestOmitemptyDropsTheEmptyOnes(t *testing.T) {
	got, err := Encode(Profile{Username: "solo"})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	for _, key := range []string{"email", "age", "nickname", "roles", "zip"} {
		if strings.Contains(got, `"`+key+`"`) {
			t.Errorf("Encode kept the empty %q: %s", key, got)
		}
	}
	for _, key := range []string{"username", "address", "city", "country", "temp"} {
		if !strings.Contains(got, `"`+key+`"`) {
			t.Errorf("Encode dropped %q, which has no omitempty: %s", key, got)
		}
	}
}

func TestNilNicknameIsDroppedButEmptyStringIsNot(t *testing.T) {
	withNil, err := Encode(Profile{Username: "a"})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if strings.Contains(withNil, "nickname") {
		t.Errorf("a nil Nickname was written: %s", withNil)
	}

	withEmpty, err := Encode(Profile{Username: "a", Nickname: ptr("")})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if !strings.Contains(withEmpty, `"nickname":""`) {
		t.Errorf("a pointer to the empty string was dropped: %s - "+
			"omitempty on a pointer tests the pointer, not what it points at", withEmpty)
	}
}

func TestEmptyRolesSliceIsDropped(t *testing.T) {
	got, err := Encode(Profile{Username: "a", Roles: []string{}})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if strings.Contains(got, "roles") {
		t.Errorf("an empty Roles slice was written: %s - omitempty covers length 0, not just nil", got)
	}
}

func TestDecode(t *testing.T) {
	in := `{"username":"ada","email":"ada@example.com","age":36,"nickname":"count",` +
		`"address":{"city":"London","zip":"E1","country":"UK"},` +
		`"roles":["admin","dev"],"temp":"21.5C"}`
	got, err := Decode([]byte(in))
	if err != nil {
		t.Fatalf("Decode returned an error: %v", err)
	}
	if got.Username != "ada" || got.Email != "ada@example.com" || got.Age != 36 {
		t.Errorf("Decode = %+v, want username ada, the email and age 36", got)
	}
	if got.Nickname == nil || *got.Nickname != "count" {
		t.Errorf("Decode Nickname = %v, want a pointer to \"count\"", got.Nickname)
	}
	if got.Address.City != "London" || got.Address.Zip != "E1" || got.Address.Country != "UK" {
		t.Errorf("Decode Address = %+v, want {London E1 UK}", got.Address)
	}
	if len(got.Roles) != 2 || got.Roles[0] != "admin" || got.Roles[1] != "dev" {
		t.Errorf("Decode Roles = %v, want [admin dev]", got.Roles)
	}
	if got.Temp != 21.5 {
		t.Errorf("Decode Temp = %v, want 21.5", float64(got.Temp))
	}
}

func TestDecodeLeavesNicknameNilWhenAbsent(t *testing.T) {
	got, err := Decode([]byte(`{"username":"ada","temp":"0C"}`))
	if err != nil {
		t.Fatalf("Decode returned an error: %v", err)
	}
	if got.Nickname != nil {
		t.Errorf("Decode Nickname = %q, want nil when the key is absent", *got.Nickname)
	}

	withEmpty, err := Decode([]byte(`{"username":"ada","nickname":"","temp":"0C"}`))
	if err != nil {
		t.Fatalf("Decode returned an error: %v", err)
	}
	if withEmpty.Nickname == nil || *withEmpty.Nickname != "" {
		t.Error(`Decode of "nickname":"" gave nil, want a pointer to the empty string - ` +
			"that is the difference the pointer buys you")
	}
}

func TestDecodeIgnoresAPasswordInTheInput(t *testing.T) {
	// "-" is a two-way rename to nothing: the field is not written, and
	// nothing in the input can reach it.
	got, err := Decode([]byte(`{"username":"ada","password":"hunter2","Password":"hunter2","temp":"0C"}`))
	if err != nil {
		t.Fatalf("Decode returned an error: %v", err)
	}
	if got.Password != "" {
		t.Errorf("Decode Password = %q, want %q - a %q tag blocks both directions", got.Password, "", "-")
	}
}

func TestDecodeRejectsABadTemperature(t *testing.T) {
	if _, err := Decode([]byte(`{"username":"ada","temp":21.5}`)); err == nil {
		t.Error("Decode of a numeric temp returned nil error, want the Celsius unmarshaller to reject it")
	}
	if _, err := Decode([]byte(`{"username":"ada","temp":"21.5F"}`)); err == nil {
		t.Error("Decode of a Fahrenheit temp returned nil error, want an error")
	}
}

func TestRoundTripKeepsEverythingItShould(t *testing.T) {
	got, err := RoundTrip(fullProfile())
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	want := fullProfile()
	if got.Username != want.Username || got.Email != want.Email || got.Age != want.Age {
		t.Errorf("RoundTrip = %+v, want the username, email and age preserved", got)
	}
	if got.Nickname == nil || *got.Nickname != *want.Nickname {
		t.Errorf("RoundTrip Nickname = %v, want a pointer to %q", got.Nickname, *want.Nickname)
	}
	if got.Address != want.Address {
		t.Errorf("RoundTrip Address = %+v, want %+v", got.Address, want.Address)
	}
	if got.Temp != want.Temp {
		t.Errorf("RoundTrip Temp = %v, want %v", float64(got.Temp), float64(want.Temp))
	}
}

func TestRoundTripDropsTheSecrets(t *testing.T) {
	got, err := RoundTrip(fullProfile())
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	if got.Password != "" {
		t.Errorf("RoundTrip Password = %q, want %q", got.Password, "")
	}
	if got.notes != "" {
		t.Errorf("RoundTrip notes = %q, want %q", got.notes, "")
	}
}

func TestRoundTripIsStable(t *testing.T) {
	once, err := RoundTrip(fullProfile())
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	first, err := Encode(once)
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	twice, err := RoundTrip(once)
	if err != nil {
		t.Fatalf("RoundTrip returned an error: %v", err)
	}
	second, err := Encode(twice)
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if first != second {
		t.Errorf("a second round trip changed the bytes:\n  %s\n  %s", first, second)
	}
}

func TestAddressTagsApplyWhereverItAppears(t *testing.T) {
	// The tags live on Address, not on the field that holds one, so encoding
	// an Address on its own gives the same keys.
	got, err := Encode(Profile{Address: Address{City: "Paris", Country: "FR"}})
	if err != nil {
		t.Fatalf("Encode returned an error: %v", err)
	}
	if !strings.Contains(got, `"address":{"city":"Paris","country":"FR"}`) {
		t.Errorf("Encode = %s, want the nested address with lowercase keys and no zip", got)
	}
}
