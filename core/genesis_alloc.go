// Copyright 2017 The go-ctereum Authors
// This file is part of the go-ctereum library.
//
// The go-ctereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ctereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ctereum library. If not, see <http://www.gnu.org/licenses/>.

package core

// Constants containing the genesis allocation of built-in genesis blocks.
// Their content is an RLP-encoded list of (address, balance) tuples.
// Use mkalloc.go to create/update them.

// nolint: misspell
const mainnetAllocData = "\xe3\xe2\x94\x1eV^\u0392\xf2\x89\x8fZ\xe4u,\xd9dX\xfa\x84I\x85\x12\x8c\x03;.<\x9f\u0400<\xe8\x00\x00\x00"
const ropstenAllocData = "\xe3\xe2\x94\x1eV^\u0392\xf2\x89\x8fZ\xe4u,\xd9dX\xfa\x84I\x85\x12\x8c\x03;.<\x9f\u0400<\xe8\x00\x00\x00"
const rinkebyAllocData = "\xe3\xe2\x94\x1eV^\u0392\xf2\x89\x8fZ\xe4u,\xd9dX\xfa\x84I\x85\x12\x8c\x03;.<\x9f\u0400<\xe8\x00\x00\x00"
const goerliAllocData = "\xf8i\xe2\x94D\x8d%@n{\x03\x1b\u01be\xd2\am\xd6\xee\u075e\xe8|I\x8c\x03;.<\x9f\u0400<\xe8\x00\x00\x00\xe2\x94\u01c1\x8d\xe7-\x9c\x88\x8e\xa0\u0425n\xb9\a\xf9\xc4\xdb\xf0'\u040c\x03;.<\x9f\u0400<\xe8\x00\x00\x00\xe2\x94\xdf\xf2\xf4\xae\xbc\xab\u02d0\xf4\xce\xf4D\xc5B\x13\x06\x10\xe3\xc8@\x8c\x03;.<\x9f\u0400<\xe8\x00\x00\x00"
const sepoliaAllocData = ""
const KilnAllocData = ""
