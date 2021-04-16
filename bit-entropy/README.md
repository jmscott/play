Here lives code to experiment with concepts of bit entropy as described in
this paper.

	https://arxiv.org/abs/1305.0954

The goal is to eventually discover a general metric for pairwise similarity
between arbitrary blobs of bits.

If we consider logical XOR (symetric difference) as the "distance" between
two bits, then the idea is to produce a 64 dimensional vector for an
arbitray blob of bits, where each component, say I, is the bit population count of the the Ith iteration of a blob derived by XORing the orginal blob.

For example, consider the 32 bit blob, say B, and POP is the population count.

	B		= 01010101  11110000  00100110  11110000
	POP(B)		= 15 bits

Now transform blob B into a 31 bit long blob by XORing adjacent bits.

	AJX(B)		= [0^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^1] [1^1] 
			  [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0] [0^0]
			  [0^0] [0^1] [1^0] [0^0] [0^1] [1^1] [1^0] [0^1]
			  [1^1] [1^1] [1^1] [1^0] [0^0] [0^0] [0^0]

			->11111110  00010000  01101010  0001000
	POP(AJX(B))	= 13 bits

and the next blob, AJX(POP(AJX(B))) yeilds a 30 bit blob

	AJX(POP(AJX(B)))= [1^1] [1^1] [1^1] [1^1] [1^1] [1^1] [1^0] [0^0]
	                  [0^0] [0^0] [0^1] [1^0] [0^0] [0^0] [0^0] [0^0]
			  [0^1] [1^1] [1^0] [0^1] [1^0] [0^1] [1^0] [0^0]
			  [0^0] [0^0] [0^1] [1^0] [0^0] [0^0]
			->00000010  00110000  10101010  001100

	POP(AJX(POP(AJX(B))))
			= 9

Stopping at three dimensions for example B, yields the bit entropy vector

	BE(B) = <15, 13, 9>

The written c code extends to 64 "dimensions", where POP(AJX(empty)) = 0
and POP(AJX(one bit blob)) = 1.

The next question is which metric defines distance in the 64 dimensional space.
