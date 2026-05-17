// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/cespare/xxhash/v2"
	"server_golang/common/types"
)

// MD5Str 计算字符串的 md5 值
func MD5Str(str string) string {
	sumBytes := md5.Sum(types.StringToBytes(str))
	return hex.EncodeToString(sumBytes[:])
}

// MD5Bytes 计算字符串的 md5 值
func MD5Bytes(b []byte) []byte {
	src := md5.Sum(b)
	dst := make([]byte, hex.EncodedLen(len(src[:])))
	hex.Encode(dst, src[:])
	return dst
}

// Hash 计算字符串的哈希值
func Hash(s []byte) uint64 {
	return xxhash.Sum64(s) // 直接对字符串哈希
}

// WXBizDataDecrypt 微信小程序加密数据解密
// 对应 PHP: WXBizDataCrypt::decryptData
// 使用 AES-128-CBC 模式，key 为 base64 解码后的 session_key，iv 为 base64 解码后的 iv
// 返回解密后的 JSON 字符串，如果解密失败返回错误
func WXBizDataDecrypt(sessionKey, encryptedData, iv string) (string, error) {
	// base64 解码 session_key
	aesKey, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return "", fmt.Errorf("decode session_key error: %v", err)
	}

	// base64 解码 iv
	aesIV, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", fmt.Errorf("decode iv error: %v", err)
	}

	// base64 解码加密数据
	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("decode encrypted_data error: %v", err)
	}

	// AES-128-CBC 解密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("aes new cipher error: %v", err)
	}

	if len(cipherText)%block.BlockSize() != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, aesIV)
	plainText := make([]byte, len(cipherText))
	mode.CryptBlocks(plainText, cipherText)

	// PKCS7 去除填充
	plainText, err = pkcs7Unpadding(plainText)
	if err != nil {
		return "", fmt.Errorf("pkcs7 unpadding error: %v", err)
	}

	return string(plainText), nil
}

// pkcs7Unpadding 移除 PKCS7 填充
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("pkcs7 unpadding: empty data")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, fmt.Errorf("pkcs7 unpadding: invalid padding size")
	}
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("pkcs7 unpadding: invalid padding byte")
		}
	}
	return data[:length-padding], nil
}
