package netscape

import (
	"marengo/modules/kooky"
	"marengo/modules/kooky/internal/cookies"
	"marengo/modules/kooky/internal/firefox/find"
	"marengo/modules/kooky/internal/netscape"
)

type netscapeFinder struct{}

var _ kooky.CookieStoreFinder = (*netscapeFinder)(nil)

func init() {
	kooky.RegisterFinder(`netscape`, &netscapeFinder{})
}

func (f *netscapeFinder) FindCookieStores() ([]kooky.CookieStore, error) {
	files, err := find.FindCookieStoreFiles(netscapeRoots, `netscape`, `cookies.txt`)
	if err != nil {
		return nil, err
	}

	var ret []kooky.CookieStore
	for _, file := range files {
		ret = append(
			ret,
			&cookies.CookieJar{
				CookieStore: &netscape.CookieStore{
					DefaultCookieStore: cookies.DefaultCookieStore{
						BrowserStr:           file.Browser,
						ProfileStr:           file.Profile,
						IsDefaultProfileBool: file.IsDefaultProfile,
						FileNameStr:          file.Path,
					},
				},
			},
		)
	}

	return ret, nil
}
