BDR Technology Justification
===================================================

There are quite a few benefits to using BDR as a basis for RDPG.

It is true in general that the lower downtime we strive for, the more sacrifices we must make in order to achieve the goal.  In the case of a top-tier, five-nines type service, it makes perfect sense that we'll need to make a few functionality compromises.  It's also important to remember, yes, we are using BDR (Bi-Directional Write Multi-Master), BUT, we're using it as if it's UDR (Uni-Directional Single Write Master).

Benefits of BDR (Should be referred to as UDR which is a subset of a full BDR implementation without the complexity, using only the advantages of BDR to achieve our goals.)
* Fast fail-over (no promotion of a slave, which is computationally expensive and error prone)
Simplified fail-over (no need switch time lines adding a new slave to the cluster, immediate)
* Easy major version upgrades
* The ability to move noisy neighbors to their own cluster easily (with the ability to add this as an automated admin feature vs support staff manually taking action on the hosts)
* Allows for transitions across globally distinct data centers with little or no downtime (near-live data and traffic migrations!)

One further point, addressing the fact that it is publicly labeled "Beta". We know for a fact that 2ndQuadrant has been running it in production for some time at several large enterprises and they did not want to remove the "Beta" label because they didn't have the bandwidth of supporting a lot of community requests. The folks at OpenSCG (after understanding the architecture and our specific usage of the product) agreed that it was a solid choice given the requirements.

We hope this helps you address the conversations around "Why are you using BDR"?!

