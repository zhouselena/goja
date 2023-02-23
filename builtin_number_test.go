package goja

import "testing"

func TestIsSafeInteger(t *testing.T) {
	const SCRIPT = `
	var maxInt = 9007199254740991
	var overflowInt = 9007199254740992
	var maxIntFloat = 9007199254740991.0
	var overflowIntInFloat = 9007199254740992.0

	assert.sameValue(Number.isSafeInteger(1.0), true);
	assert.sameValue(Number.isSafeInteger(1), true);
	assert.sameValue(Number.isSafeInteger(0), true);
	assert.sameValue(Number.isSafeInteger(-1), true);
	assert.sameValue(Number.isSafeInteger(maxInt), true);
	assert.sameValue(Number.isSafeInteger(maxIntFloat), true);
	assert.sameValue(Number.isSafeInteger(overflowInt), false);
	assert.sameValue(Number.isSafeInteger(overflowIntInFloat), false);
	assert.sameValue(Number.isSafeInteger('1'), false);
	assert.sameValue(Number.isSafeInteger(1.1), false);
	assert.sameValue(Number.isSafeInteger(Math.abs(1.0)), true);
	assert.sameValue(Number.isSafeInteger(Math.abs(1.1)), false);
	`
	testScript(TESTLIB+SCRIPT, _undefined, t)
}
