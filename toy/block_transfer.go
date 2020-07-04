package toy

import "github.com/tchajed/goose/machine/disk"

// TransferEvenBlock assumes it is given ownership of a and that a initially has
// an even block (defined as the first byte being even).
//
// The spec is that TransferEvenBlock preserves that a has an even block (across
// crashes) and is safe (that is, the panic does not get triggered)
func TransferEvenBlock(d disk.Disk, a uint64) {
	// create a crash invariant for a

	// logically transfer a to this function
	go func() {
		b4 := make(disk.Block, disk.BlockSize)
		b4[0] = 4
		d.Write(a, b4)
		b := d.Read(a)
		if b[0] != 4 {
			// the proof will show this does not happen (which would otherwise
			// get stuck in the semantics)
			panic("unexpected value on disk")
		}
	}()
}
