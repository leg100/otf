package releases

//func Test_latestChecker(t *testing.T) {
//	tests := []struct {
//		name string
//		last time.Time // last time checked
//		got  string    // version returned
//	}{
//		{"skip check", time.Now(), ""},
//		{"perform check", time.Time{}, "1.6.1"},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// endpoint is a stub endpoint that always returns 1.6.1 as latest
//			// version
//			endpoint := func() string {
//				mux := http.NewServeMux()
//				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//					w.Header().Add("Content-Type", "application/json")
//					w.Write(testutils.ReadFile(t, "./testdata/latest.json"))
//				})
//				srv := httptest.NewServer(mux)
//				t.Cleanup(srv.Close)
//				u, err := url.Parse(srv.URL)
//				require.NoError(t, err)
//				return u.String()
//			}()
//
//			v, err := latestChecker{endpoint}.check(tt.last)
//			require.NoError(t, err)
//			assert.Equal(t, tt.got, v)
//		})
//	}
//}
