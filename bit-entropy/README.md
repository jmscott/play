Here lives code to explore the concept of bit entropy as described in
this paper.

[BiEntropy - The Approximate Entropy of a Finite Binary String]
(https://arxiv.org/abs/1305.0954)

The goal is to discover a general metric for pairwise similarity between
arbitrary blobs of bits.

If we consider `logical XOR` as the "distance" between two bits, then the idea
is to produce a 64 dimensional vector for an arbitray blob of bits, where each
dimension, say `I`, is the bit population count of the the `Ith` iteration of
a blob derived by `XORing` adjacents bits.

For example, consider a 32 bit blob, say `B`, `POP`, where `POP` is the
population count and `^` is the `C` language bitwise xor.

```

    B               =  01010101  11110000  00100110  11110000

    POP(B)          =  15 bits
```

Now transform blob B into a 31 bit long blob by `XORing` adjacent bits.

```
    AJX(B)          =  [0^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^1] [1^1] 
                       [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0] [0^0]
                       [0^0] [0^1] [1^0] [0^0] [0^1] [1^1] [1^0] [0^1]
                       [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0]

                    => 11111110  00010000  01101011  0001000

    POP(AJX(B))     =  13 bits
```

and, again, AJX(AJX(B)) yields a 30 bit blob

```
    AJX(AJX(B)))    =  [1^1] [1^1] [1^1] [1^1] [1^1] [1^1] [1^0] [0^0]
                       [0^0] [0^0] [0^1] [1^0] [0^0] [0^0] [0^0] [0^0]
                       [0^1] [1^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^0]
                       [0^0] [0^0] [0^1] [1^0] [0^0] [0^0]

                    => 00000010  00110000  10101010  001100

    POP(AJX(AJX(B)))
                    =  9 bits
```

and to the third power, (AJX^3)(B) yields a 29 bit vector

    (AJX^3)(B)      =  [0^0] [0^0] [0^0] [0^0] [0^0] [0^1] [1^0] [0^0]
                       [0^0] [0^1] [1^1] [1^0] [0^0] [0^0] [0^0] [0^1]
		       [1^0] [0^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^0]
		       [0^0] [0^1] [1^1] [1^0] [0^0] 

		    => 00000110  01010001  11111110  01010


    POP((AJX^3)(B)) =  14

and to the fourth power, (AJX^4)(B) yields a 28 bit vector

    (AJX^4)(B)      =   [0^0] [0^0] [0^0] [0^0] [0^1] [1^1] [1^0] [0^0]
                        [0^1] [1^0] [0^1] [1^0] [0^0] [0^0] [0^1] [1^1]
                        [1^1] [1^1] [1^1] [1^1] [1^1] [1^1] [1^0] [0^0]
                        [0^1] [1^0] [0^1] [1^0]
                    
                    =>  00001010  11110010  00000010  1111

    POP((AJX^4)(B)) =   12

and to the fifth power, (AJX^5)(B) yields a 27 bit vector

    (AJX^5)(B)      =   [0^0] [0^0] [0^0] [0^1] [1^0] [0^1] [1^0] [0^1]
                        [1^1] [1^1] [1^1] [1^0] [0^0] [0^1] [1^0] [0^0]
			[0^0] [0^0] [0^0] [0^0] [0^0] [0^1] [1^0] [0^1]
			[1^1] [1^1] [1^1]

                    =>  00011111  00010110  00000111  000

    POP((AJX^5)(B)) =   11

Stopping at 6 dimensions for `B`, yields the vector

```
        BEV(B)       =  <15 bits, 13 bits, 9 bits, 14 bits, 12bits, 11bits>
```

The written `C` code extends to 64 "dimensions", where `POP(AJX(empty)) = 0`
and `POP(AJX(one bit blob)) = 1`.

The next question is which metric on `BEM` defines a notion of distance in the
64 dimensional space.
