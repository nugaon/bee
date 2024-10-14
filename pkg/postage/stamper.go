// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postage

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/storage"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

var (
	// ErrBucketFull is the error when a collision bucket is full.
	ErrBucketFull = errors.New("bucket full")
)

// Stamper can issue stamps from the given address of chunk.
type Stamper interface {
	Stamp(swarm.Address) (*Stamp, error)
	BatchId() []byte
}

// stamper connects a stampissuer with a signer.
// A stamper is created for each upload session.
type stamper struct {
	store  storage.Store
	issuer *StampIssuer
	signer crypto.Signer
}

// NewStamper constructs a Stamper.
func NewStamper(store storage.Store, issuer *StampIssuer, signer crypto.Signer) Stamper {
	return &stamper{store, issuer, signer}
}

// Stamp takes chunk, see if the chunk can be included in the batch and
// signs it with the owner of the batch of this Stamp issuer.
func (st *stamper) Stamp(addr swarm.Address) (*Stamp, error) {
	st.issuer.mtx.Lock()
	defer st.issuer.mtx.Unlock()

	item := &StampItem{
		BatchID:      st.issuer.data.BatchID,
		chunkAddress: addr,
	}
	lockKey := fmt.Sprintf("postageIdStamp-%x", st.issuer.data.BatchID)
	fmt.Printf("stamper.go: Stamp: Locking StampLocker %s\n", lockKey)
	StampLocker.Lock(lockKey)
	defer StampLocker.Unlock(lockKey)
	fmt.Printf("stamper.go: Stamp: Locking StampLocker LOCKED!!!!!!!!!!! %s\n", lockKey)

	err := st.store.Get(item)
	fmt.Printf("stamper.go: Stamp: err: %v\n", err)

	if errors.Is(err, storage.ErrNotFound) {
		bucket := ToBucket(st.issuer.BucketDepth(), addr)
		item.BatchIndex = indexToBytes(bucket, 0)
		item.BatchTimestamp = unixTime()
		if err = st.store.Put(item); err != nil {
			fmt.Printf("maybe does it go here?????????????????????? %s\n", err)
			return nil, err
		}
		fmt.Printf("all good with stamp %s\n", item)
	} else if err == nil {
		fmt.Printf("unknown territory semmingly$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$....\n")
		item.BatchIndex, item.BatchTimestamp, err = st.issuer.increment(addr)
		if err != nil {
			return nil, err
		}
		if err := st.store.Put(item); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("get stamp for %s: %w", item, err)
	}

	toSign, err := ToSignDigest(
		addr.Bytes(),
		st.issuer.data.BatchID,
		item.BatchIndex,
		item.BatchTimestamp,
	)
	if err != nil {
		return nil, err
	}
	sig, err := st.signer.Sign(toSign)
	if err != nil {
		return nil, err
	}
	fmt.Printf("stamper.go: Stamp: returning NewStamp(%x, %x, %x, %x)\n", st.issuer.data.BatchID, item.BatchIndex, item.BatchTimestamp, sig)
	return NewStamp(st.issuer.data.BatchID, item.BatchIndex, item.BatchTimestamp, sig), nil
}

// BatchId gives back batch id of stamper
func (st *stamper) BatchId() []byte {
	return st.issuer.data.BatchID
}

type presignedStamper struct {
	stamp *Stamp
	owner []byte
}

func NewPresignedStamper(stamp *Stamp, owner []byte) Stamper {
	return &presignedStamper{stamp, owner}
}

func (st *presignedStamper) Stamp(addr swarm.Address) (*Stamp, error) {
	// check stored stamp is against the chunk address
	// Recover the public key from the signature
	signerAddr, err := RecoverBatchOwner(addr, st.stamp)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(st.owner, signerAddr) {
		return nil, ErrInvalidBatchSignature
	}

	return st.stamp, nil
}

func (st *presignedStamper) BatchId() []byte {
	return st.stamp.BatchID()
}
