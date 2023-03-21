package rlwe

// EvaluationKey is a public key indended to be used during the evaluation phase of a homomorphic circuit.
// It provides a one way public and non-interactive re-encryption from a ciphertext encrypted under `skIn`
// to a ciphertext encrypted under `skOut`.
//
// Such re-encryption is for example used for:
//
// - Homomorphic relinearization: re-encryption of a quadratic ciphertext (that requires (1, sk sk^2) to be decrypted)
// to a linear ciphertext (that required (1, sk) to be decrypted). In this case skIn = sk^2 an skOut = sk.
//
// - Homomorphic automorphisms: an automorphism in the ring Z[X]/(X^{N}+1) is defined as pi_k: X^{i} -> X^{i^k} with
// k coprime to 2N. Pi_sk is for exampled used during homomorphic slot rotations. Applying pi_k to a ciphertext encrypted
// under sk generates a new ciphertext encrypted under pi_k(sk), and an Evaluationkey skIn = pi_k(sk) to skOut = sk
// is used to bring it back to its original key.
type EvaluationKey struct {
	GadgetCiphertext
}

// NewEvaluationKey returns a new EvaluationKey with pre-allocated zero-value
func NewEvaluationKey(params Parameters, levelQ, levelP int) *EvaluationKey {
	return &EvaluationKey{GadgetCiphertext: *NewGadgetCiphertext(
		params,
		levelQ,
		levelP,
		params.DecompRNS(levelQ, levelP),
		params.DecompPw2(levelQ, levelP),
	)}
}

// Equals checks two EvaluationKeys for equality.
func (evk *EvaluationKey) Equals(other *EvaluationKey) bool {
	return evk.GadgetCiphertext.Equals(&other.GadgetCiphertext)
}

// CopyNew creates a deep copy of the target EvaluationKey and returns it.
func (evk *EvaluationKey) CopyNew() *EvaluationKey {
	return &EvaluationKey{GadgetCiphertext: *evk.GadgetCiphertext.CopyNew()}
}

// MarshalBinarySize returns the size in bytes that the object once marshalled into a binary form.
func (evk *EvaluationKey) MarshalBinarySize() (dataLen int) {
	return evk.GadgetCiphertext.MarshalBinarySize()
}

// MarshalBinary encodes the object into a binary form on a newly allocated slice of bytes.
func (evk *EvaluationKey) MarshalBinary() (data []byte, err error) {
	data = make([]byte, evk.MarshalBinarySize())
	_, err = evk.Read(data)
	return
}

// Read encodes the object into a binary form on a preallocated slice of bytes
// and returns the number of bytes written.
func (evk *EvaluationKey) Read(data []byte) (ptr int, err error) {
	return evk.GadgetCiphertext.Read(data)
}

// UnmarshalBinary decodes a slice of bytes generated by MarshalBinary
// or Read on the object.
func (evk *EvaluationKey) UnmarshalBinary(data []byte) (err error) {
	_, err = evk.Write(data)
	return
}

// Write decodes a slice of bytes generated by MarshalBinary or
// Read on the object and returns the number of bytes read.
func (evk *EvaluationKey) Write(data []byte) (ptr int, err error) {
	return evk.GadgetCiphertext.Write(data)
}
