# Example programs for Perennial

[![Build Status](https://travis-ci.com/mit-pdos/perennial-examples.svg?branch=master)](https://travis-ci.com/mit-pdos/perennial-examples)

## dir - single directory

Manages a fixed number of inodes, which support block appends and reads.

Illustrates the shadow update pattern. To append to an inode, we allocate a
block, update it with full ownership, and then atomically install. However,
while updating we still have a crash obligation to the allocator.

Illustrates modularity. We're able to structure this into three components: an
allocator, an inode, and a directory. The allocator an inode are completely
independent.

* The allocator is an in-memory structure that serializes access to the free
  space. It has to be restored during recovery by figuring out what has been
  allocated.
* The inode is a durable, append-only structure. It requires the caller to
  prepare and supply disk blocks to append.
* The directory composes multiple inodes and an allocator.

Note that the code doesn't literally separate these into Go packages (for our
convenience, though maybe we should).
