#!/usr/bin/env murex

a [0..255] -> foreach $i {
    out "{ESC}[38:5:$(i)m$i"
}