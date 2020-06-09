# Example programs for Perennial

[![Build Status](https://travis-ci.com/mit-pdos/perennial-examples.svg?branch=master)](https://travis-ci.com/mit-pdos/perennial-examples)

## dir - single directory

Manages a fixed number of inodes, which support block appends and reads.

Illustrates the shadow update pattern. To append to an inode, we allocate a
block, update it with full ownership, and then atomically install. However,
while updating we still have a crash obligation to the allocator.

Illustrates modularity. We're able to structure this into three components: an
allocator, an inode, and a directory. The inode uses the allocator but does not
manage it (it does not restore the allocator, and shares it).

* The [allocator](alloc/) is an in-memory structure that serializes access to the free
  space. It has to be restored during recovery by figuring out what has been
  allocated.
* The [inode](inode/) is a durable, append-only structure. It calls into an
  allocator which is shared among all the inodes.
* The [directory](dir/) composes multiple inodes and an allocator.
* The [single-inode package](single_inode/) is a simple client of the inode; as
  an alternative to the directory, it only has a single inode.

## replicated block

Self-contained example at [replicated_block](replicated_block/). Intended to
resemble the Perennial replicated disk example.

Toy example that replicates a single logical disk block across two blocks. Reads
can pick which replica to use, while writes update both. For consistency, we
lock the disk addresses and recovery by syncing. This requires a crash lock, and
its crash invariant allows the addresses to be out-of-sync.

This is a modular proof that can be instantiated by the caller for multiple
addresses, building a replicated disk out of tiny pieces.

Note that this finally takes the 6.826 example and makes it horizontally
modular, vertically composable, concurrent, and written in Go.

## append-only log

(needs to be moved from Goose repo to here)

Implements atomically appending blocks to a log using a header block.

## circular buffer

(needs to be re-implemented)

Similar to the append-only log, but supports concurrent logging and installation
using a circular on-disk structure. More realistically than the append-only log,
the only read operation is recovery, since the caller is expected to be managing
a cache anyway.
