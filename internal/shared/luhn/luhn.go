// Package luhn implements Luhn algorithm validation for order numbers.
package luhn

// Valid returns true if the string s passes the Luhn check.
func Valid(s string) bool {
	if len(s) < 2 {
		return false
	}

	var sum int
	parity := len(s) % 2

	for i, c := range s {
		if c < '0' || c > '9' {
			return false
		}

		d := int(c - '0')

		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}
