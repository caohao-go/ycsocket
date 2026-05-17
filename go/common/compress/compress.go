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

package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"sync"

	"server_golang/common/json"
)

var (
	gzipWriterPool = sync.Pool{}
	gzipReaderPool = sync.Pool{}
)

// JsonMarshalAndCompress first json encoding, then compression.
func JsonMarshalAndCompress(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}

	gzipWriter, ok := gzipWriterPool.Get().(*gzip.Writer)
	if !ok {
		gzipWriter = gzip.NewWriter(buf)
	} else {
		gzipWriter.Reset(buf)
	}

	defer gzipWriterPool.Put(gzipWriter)

	if err := json.Api.NewEncoder(gzipWriter).Encode(v); err != nil {
		_ = gzipWriter.Close()
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressJsonUnmarshal decompression and json decode
func DecompressJsonUnmarshal(in []byte, v interface{}) error {
	gzipReader, ok := gzipReaderPool.Get().(*gzip.Reader)

	defer func() {
		if gzipReader != nil {
			gzipReaderPool.Put(gzipReader)
		}
	}()

	var err error
	if ok {
		err = gzipReader.Reset(bytes.NewReader(in))
	} else {
		gzipReader, err = gzip.NewReader(bytes.NewReader(in))
	}

	var reader io.Reader = gzipReader
	if err != nil {
		reader = bytes.NewReader(in)
	}

	return json.Api.NewDecoder(reader).Decode(v)
}

func Decompress(in []byte) ([]byte, error) {
	gzipReader, ok := gzipReaderPool.Get().(*gzip.Reader)

	defer func() {
		if gzipReader != nil {
			gzipReaderPool.Put(gzipReader)
		}
	}()

	var err error
	if ok {
		err = gzipReader.Reset(bytes.NewReader(in))
	} else {
		gzipReader, err = gzip.NewReader(bytes.NewReader(in))
	}

	var reader io.Reader = gzipReader
	if err != nil {
		reader = bytes.NewReader(in)
	}

	return ioutil.ReadAll(reader)
}
