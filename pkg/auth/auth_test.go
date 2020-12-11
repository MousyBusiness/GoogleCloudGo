package auth

import "testing"

func TestCheckInternal(t *testing.T) {
	// 10.0.0.0
	where := checkInternal("10.0.0.2")
	want := true
	if where != want {
		t.Errorf("ip 10.0.0.2, wanted: %v, got: %v", want, where)
	}

	// 127.0.0.0
	where = checkInternal("127.0.0.1")
	want = true
	if where != want {
		t.Errorf("ip 127.0.0.1, wanted: %v, got: %v", want, where)
	}

	// 172.16
	where = checkInternal("172.16.0.1")
	want = true
	if where != want {
		t.Errorf("ip 172.16.0.1, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("172.0.0.1")
	want = false
	if where != want {
		t.Errorf("ip 172.0.0.1, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("172.32.0.1")
	want = false
	if where != want {
		t.Errorf("ip 172.32.0.1, wanted: %v, got: %v", want, where)
	}

	// 192.168
	where = checkInternal("192.168.0.1")
	want = true
	if where != want {
		t.Errorf("ip 192.168.0.1, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("192.167.0.1")
	want = false
	if where != want {
		t.Errorf("ip 192.167.0.1, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("192.169.0.1")
	want = false
	if where != want {
		t.Errorf("ip 192.169.0.1, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("192.168.234.1")
	want = true
	if where != want {
		t.Errorf("ip 192.168.234.1, wanted: %v, got: %v", want, where)
	}

	// externals
	where = checkInternal("2.32.0.222")
	want = false
	if where != want {
		t.Errorf("ip 2.32.0.222, wanted: %v, got: %v", want, where)
	}

	where = checkInternal("32.18.1.23")
	want = false
	if where != want {
		t.Errorf("32.18.1.23, wanted: %v, got: %v", want, where)
	}
}
