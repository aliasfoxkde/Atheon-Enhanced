// bundle_test_utils.go provides test helpers for bundle-related tests.
// It is compiled as part of the core test binary.
package core

// testSetBundleURL calls SetBundleDownloadURL with skipHostValidation enabled so
// tests can point at httptest servers on loopback without triggering the SSRF guard.
func testSetBundleURL(url string) func() {
	skipHostValidation = true
	restore := SetBundleDownloadURL(url)
	return func() {
		restore()
		skipHostValidation = false
	}
}
