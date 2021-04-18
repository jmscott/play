# Overview

Here lives code to explore the concept of bit entropy as described in
the Arxiv paper.

[BiEntropy - The Approximate Entropy of a Finite Binary String]
(https://arxiv.org/abs/1305.0954)

The goal is to discover a general metric for pairwise similarity between
arbitrary blobs of bits.  Two blobs will be similar if their entropy is
similar and low, possibly also considering other core attributes, such as byte
count and bit population.

## Derivative of Bit String `B`

If we consider `logical XOR` as the "distance" between two bits, then consider
the "derivative" of a bit string, N bits long, as the XOR of adjacent bits,
yielding another bit string N-1 bits long.

For example, consider a 32 bit blob, say `B`, `POP`, where `POP` is the
population count and `^` is the `C` language bitwise `XOR`.
```
    B               =  01010101  11110000  00100110  11110000

    POP(B)          =  15 bits
```

### First Derivative of `B`

Now "differentiate" blob `B` into a 31 bit blob by `XORing` adjacent bits.
```
    BDV(B)          =  [0^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^1] [1^1] 
                       [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0] [0^0]
                       [0^0] [0^1] [1^0] [0^0] [0^1] [1^1] [1^0] [0^1]
                       [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0]

                    => 11111110  00010000  01101011  0001000

    POP(BDV(B))     =  13 bits
```

### Second Derivative of `B`

and, again, the second derivative, `BDV(BDV(B))`, yields a 30 bit blob

```
    BDV(BDV(B)))    =  [1^1] [1^1] [1^1] [1^1] [1^1] [1^1] [1^0] [0^0]
                       [0^0] [0^0] [0^1] [1^0] [0^0] [0^0] [0^0] [0^0]
                       [0^1] [1^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^0]
                       [0^0] [0^0] [0^1] [1^0] [0^0] [0^0]

                    => 00000010  00110000  10101010  001100

    POP(BDV(BDV(B)))
                    =  9 bits
```

### Third Derivative of `B`

and to the third power, (BDV^3)(B) yields a 29 bit blob
```
    (BDV^3)(B)      =  [0^0] [0^0] [0^0] [0^0] [0^0] [0^1] [1^0] [0^0]
                       [0^0] [0^1] [1^1] [1^0] [0^0] [0^0] [0^0] [0^1]
                       [1^0] [0^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^0]
                       [0^0] [0^1] [1^1] [1^0] [0^0] 

                    => 00000110  01010001  11111110  01010

    POP((BDV^3)(B)) =  14
```

### Fourth Derivative of `B`

and to the fourth power, (BDV^4)(B), yields a 28 bit blob
```
    (BDV^4)(B)      =   [0^0] [0^0] [0^0] [0^0] [0^1] [1^1] [1^0] [0^0]
                        [0^1] [1^0] [0^1] [1^0] [0^0] [0^0] [0^1] [1^1]
                        [1^1] [1^1] [1^1] [1^1] [1^1] [1^1] [1^0] [0^0]
                        [0^1] [1^0] [0^1] [1^0]
                    
                    =>  00001010  11110010  00000010  1111

    POP((BDV^4)(B)) =   12
```

### Fifth  Derivative of `B`

and to the fifth power, `(BDV^5)(B)`, yields a 27 bit blob
```
    (BDV^5)(B)      =   [0^0] [0^0] [0^0] [0^1] [1^0] [0^1] [1^0] [0^1]
                        [1^1] [1^1] [1^1] [1^0] [0^0] [0^1] [1^0] [0^0]
                        [0^0] [0^0] [0^0] [0^0] [0^0] [0^1] [1^0] [0^1]
                        [1^1] [1^1] [1^1]

                    =>  00011111  00010110  00000111  000

    POP((BDV^5)(B)) =   11
```

##  Sequence of Bit Population (Hamming Distance)

Stopping at 5 dimensions for `B`, yields the sequence of bit counts

```
        BEV(B)       =  [13 bits, 9 bits, 14 bits, 12 bits, 11 bits]
```

Eventually the sequence converges to either `0` or `1`.

# Links

- [BiEntropy - The Approximate Entropy of a Finite Binary String](https://arxiv.org/abs/1305.0954)
- [Hamming Distance](https://en.wikipedia.org/wiki/Hamming_distance)
- [String Metric](https://en.wikipedia.org/wiki/String_metric)
