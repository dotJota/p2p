#!/usr/bin/python                                                                            
    
import sys
from subprocess import call
                                                                                             
from mininet.topo import Topo
from mininet.net import Mininet
from mininet.util import dumpNodeConnections
from mininet.log import setLogLevel, info
from mininet.link import TCLink

from generateTreeTopo import createTreeTopo

class CustomTopo(Topo):
	"Creates a customized topology based on net_topo."
	def build(self, net_topo):
		
		for net in net_topo['networks'].keys():
			self.addSwitch(net)
			for net_ in net_topo['networks'][net].keys():
				args = net_topo['networks'][net][net_]
				self.addLink(net, net_,**args)
		
		n = len(net_topo['peers'].keys())
		for peer in net_topo['peers'].keys():
			self.addHost(peer)
			for net_ in net_topo['peers'][peer].keys():
				args = net_topo['peers'][peer][net_]
				self.addLink(peer, net_,**args)
		
def connectionTest(nHosts):

	"Create and test a tree network. The degree of each node is limited to 20."
	
	fullTopo = createTreeTopo(nHosts, max_children=2)
	topo = CustomTopo(fullTopo)
	
	try:
		net = Mininet(topo, link=TCLink)
		net.start()
		
		print( "Dumping switch connections" )
		dumpNodeConnections(net.switches)

		print( "Setting up nodes" )
		hosts = net.hosts
		for i in range(nHosts):			
			# node execution arguments
			nPeers = "--n " + str(nHosts)
			pIndex = "--i " + str(i)
			baseIP = "--base_ip " + "10.0.0.1"
			port = "--base_port " + "5001"
			
			cmd = "../peer/peer " + nPeers + " " + pIndex + " " + baseIP + " " + port + " 2>&1"
			hosts[i].sendCmd(cmd, shell=True)
		
		for h in hosts:
			print(h.waitOutput())
		
		net.stop()
		
	finally:
		call("pkill -f peer", shell = True)
		call(["mn","-c"])
		print("The end")

if __name__ == '__main__':
	"""
		Usage: sudo python3 launchTree_test.py number_of_Hosts
	"""
	# Tell mininet to print useful information
	setLogLevel('info')
	connectionTest(int(sys.argv[1]))
