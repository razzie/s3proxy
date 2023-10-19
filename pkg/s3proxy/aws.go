package s3proxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func GetAWSAccessKey(r *http.Request) (accessKey string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}

	// Check if SigV2
	if strings.HasPrefix(authHeader, "AWS ") {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			return "", nil
		}
		creds := strings.Split(parts[1], ":")
		if len(creds) != 2 {
			return "", nil
		}
		return creds[0], nil
	}

	// Check if SigV4
	if strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 ") {
		// Parse the Authorization header for SigV4
		re := regexp.MustCompile(`Credential=([^,]+),`)
		matches := re.FindStringSubmatch(authHeader)
		if len(matches) != 2 {
			return "", nil
		}
		creds := strings.Split(matches[1], "/")
		if len(creds) != 2 {
			return "", nil
		}
		return creds[0], nil
	}

	return "", nil
}

func CalculateSignatureV2(accessKey, secretKey, method, contentType, contentMD5, date, canonicalizedResource string, headers http.Header) (string, string) {
	var stringToSign string
	stringToSign = strings.Join([]string{
		method,
		contentMD5,
		contentType,
		date,
	}, "\n") + "\n"

	// Add canonicalized headers to the stringToSign
	var headerKeys []string
	for key := range headers {
		key = strings.ToLower(key)
		if strings.HasPrefix(key, "x-amz-") {
			headerKeys = append(headerKeys, key)
		}
	}

	sort.Strings(headerKeys)
	for _, key := range headerKeys {
		stringToSign += fmt.Sprintf("%s:%s\n", key, headers.Get(key))
	}

	// Add canonicalized resource to the stringToSign
	stringToSign += canonicalizedResource

	// Calculate the signature
	hmacSha1 := hmac.New(sha1.New, []byte(secretKey))
	hmacSha1.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(hmacSha1.Sum(nil))

	return stringToSign, signature
}
