// Package doctor implements the diagnostics behind `secwan doctor`:
// environment, manifest, key, recipient, storage, and profile checks.
package doctor

// Run executes the full diagnostic suite and exits non-zero when any check
// fails.
func Run() {
	run()
}
