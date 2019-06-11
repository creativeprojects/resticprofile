#/bin/sh

diskutil erasevolume HFS+ RAMDisk `hdiutil attach -nomount ram://4194304`
