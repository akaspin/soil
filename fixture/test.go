package fixture

//import (
//	"os"
//	"testing"
//)
//
//// Run test only if certain OS Variables is defined
//func RunTestIf(t *testing.T, env ...string) {
//	t.Helper()
//	for _, v := range env {
//		if _, ok := os.LookupEnv(v); !ok {
//			t.Skipf("skipping: %s is not defined", v)
//		}
//	}
//}
//
//// Run test only if certain OS Variables is not defined
//func RunTestUnless(t *testing.T, env ...string) {
//	t.Helper()
//	for _, v := range env {
//		if _, ok := os.LookupEnv(v); ok {
//			t.Skipf("skipping: %s is defined", v)
//		}
//	}
//}
