package ckks

import (
	"fmt"
	"math/big"

	"github.com/tuneinsight/lattigo/v4/ring"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/tuneinsight/lattigo/v4/rlwe/ringqp"
	"github.com/tuneinsight/lattigo/v4/utils"
	"github.com/tuneinsight/lattigo/v4/utils/bignum"
)

// LinearTransformEncoder is a struct complying to the rlwe.LinearTransformEncoder.
type LinearTransformEncoder[T float64 | complex128 | *big.Float | *bignum.Complex] struct {
	*Encoder
	diagonals map[int][]T
	values    []T
}

// NewLinearTransformEncoder creates a new LinearTransformEncoder.
func NewLinearTransformEncoder[T float64 | complex128 | *big.Float | *bignum.Complex](ecd *Encoder, diagonals map[int][]T) rlwe.LinearTransformEncoder {
	return LinearTransformEncoder[T]{
		Encoder:   ecd,
		diagonals: diagonals,
		values:    make([]T, ecd.Parameters().MaxSlots()),
	}
}

// Parameters returns the rlwe.Parameters of the underlying LinearTransformEncoder.
func (l LinearTransformEncoder[_]) Parameters() rlwe.Parameters {
	return l.Encoder.Parameters().Parameters
}

// NonZeroDiagonals returns the list of non-zero diagonals of the matrix stored in the underlying LinearTransformEncoder.
func (l LinearTransformEncoder[_]) NonZeroDiagonals() []int {
	return utils.GetKeys(l.diagonals)
}

// EncodeLinearTransformDiagonalNaive encodes the i-th non-zero diagonal of the internaly stored matrix at the given scale on the outut polynomial.
func (l LinearTransformEncoder[_]) EncodeLinearTransformDiagonalNaive(i int, scale rlwe.Scale, LogSlots int, output ringqp.Poly) (err error) {

	if diag, ok := l.diagonals[i]; ok {
		return l.Embed(diag, LogSlots, scale, true, output)
	}

	return fmt.Errorf("cannot EncodeLinearTransformDiagonalNaive: diagonal [%d] doesn't exist", i)
}

// EncodeLinearTransformDiagonal encodes the i-th non-zero diagonal  of size at most 2^{LogSlots} rotated by `rot` positions
// to the left of the internaly stored matrix at the given Scale on the outut ringqp.Poly.
func (l LinearTransformEncoder[T]) EncodeLinearTransformDiagonal(i, rot int, scale rlwe.Scale, logSlots int, output ringqp.Poly) (err error) {

	ecd := l.Encoder
	slots := 1 << logSlots

	// manages inputs that have rotation between 0 and slots-1 or between -slots/2 and slots/2-1
	v, ok := l.diagonals[i]
	if !ok {
		v = l.diagonals[i-slots]
	}

	rot &= (slots - 1)

	var values []T
	if rot != 0 {

		values = l.values

		if slots >= rot {
			copy(values[:slots-rot], v[rot:])
			copy(values[slots-rot:], v[:rot])
		} else {
			copy(values[slots-rot:], v)
		}
	} else {
		values = v[:slots]
	}

	return ecd.Embed(values[:slots], logSlots, scale, true, output)
}

// TraceNew maps X -> sum((-1)^i * X^{i*n+1}) for 0 <= i < N and returns the result on a new ciphertext.
// For log(n) = logSlots.
func (eval *Evaluator) TraceNew(ctIn *rlwe.Ciphertext, logSlots int) (ctOut *rlwe.Ciphertext) {
	ctOut = NewCiphertext(eval.params, 1, ctIn.Level())
	eval.Trace(ctIn, logSlots, ctOut)
	return
}

// Average returns the average of vectors of batchSize elements.
// The operation assumes that ctIn encrypts SlotCount/'batchSize' sub-vectors of size 'batchSize'.
// It then replaces all values of those sub-vectors by the component-wise average between all the sub-vectors.
// Example for batchSize=4 and slots=8: [{a, b, c, d}, {e, f, g, h}] -> [0.5*{a+e, b+f, c+g, d+h}, 0.5*{a+e, b+f, c+g, d+h}]
// Operation requires log2(SlotCout/'batchSize') rotations.
// Required rotation keys can be generated with 'RotationsForInnerSumLog(batchSize, SlotCount/batchSize)”
func (eval *Evaluator) Average(ctIn *rlwe.Ciphertext, logBatchSize int, ctOut *rlwe.Ciphertext) {

	if ctIn.Degree() != 1 || ctOut.Degree() != 1 {
		panic("ctIn.Degree() != 1 or ctOut.Degree() != 1")
	}

	if logBatchSize > ctIn.LogSlots {
		panic("cannot Average: batchSize must be smaller or equal to the number of slots")
	}

	ringQ := eval.params.RingQ()

	level := utils.Min(ctIn.Level(), ctOut.Level())

	n := 1 << (ctIn.LogSlots - logBatchSize)

	// pre-multiplication by n^-1
	for i, s := range ringQ.SubRings[:level+1] {

		invN := ring.ModExp(uint64(n), s.Modulus-2, s.Modulus)
		invN = ring.MForm(invN, s.Modulus, s.BRedConstant)

		s.MulScalarMontgomery(ctIn.Value[0].Coeffs[i], invN, ctOut.Value[0].Coeffs[i])
		s.MulScalarMontgomery(ctIn.Value[1].Coeffs[i], invN, ctOut.Value[1].Coeffs[i])
	}

	eval.InnerSum(ctOut, 1<<logBatchSize, n, ctOut)
}
