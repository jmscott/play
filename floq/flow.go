package main

//  a river to my people ...
type flow_chan chan *flow

type flow struct {

        resolved chan struct{}

	next chan flow_chan
}

func (flo *flow) get() *flow {

        <-flo.resolved

        //  next active flow arrives on this channel
        reply := make(flow_chan)

        //  request another flow, sending reply channel to mother
        flo.next <- reply

        //  return next flow
        return <-reply
}
