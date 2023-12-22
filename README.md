Dynamic JitterBuffer - A JitterBuffer buffers incoming packets and makes sure they are in order. If a packet was lost in transmission it also allows the developer to respond as they wish (NACK, PLI….)

Pion currently only has a static JitterBuffer. You set a static number of packets you want to wait. Pion can then re-order and send NACKs. This doesn’t work well for diverse networks. This could add latency to users with great networks. Inversely users with bad networks might not have enough delay.

