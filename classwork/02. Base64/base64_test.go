package main

import (
    "testing"
    "encoding/base64"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
    if a != b {
        t.Fail()
    }
}

func testBase64ByDecoding(t *testing.T, str string) {
    data, _ := base64.StdEncoding.DecodeString(base64encode(str))
    assertEqual(t, string(data), str)
}

func TestBase64(t *testing.T) {
	testBase64ByDecoding(t, "")
	testBase64ByDecoding(t, "a")
	testBase64ByDecoding(t, "ab")
	testBase64ByDecoding(t, "xyz")
	testBase64ByDecoding(t, "xyzw")
	testBase64ByDecoding(t, "aaaaa")
	testBase64ByDecoding(t, "aaaaaaaa")
	testBase64ByDecoding(t, "aaaaaaaaaaaaaaaa")
	testBase64ByDecoding(t, "aaaaaaaaaaaaaaaaa")
	testBase64ByDecoding(t, "aaaaaabaaaaaaaaabaaaaaaaaaaaaaaaaaaba=")
	testBase64ByDecoding(t, "^&^UHASFT^&Husanfus89&^ASYdhhewd6IFBHASFGsf-=-0-9292370(#*)(^&%&$@#*)")
}
