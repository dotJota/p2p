#!/usr/bin/python                                                                            
     
from time import time                   
from time import sleep        
from signal import SIGINT
from subprocess import call
import json
import sys
                                                                                             
from mininet.topo import Topo
from mininet.net import Mininet
from mininet.util import dumpNodeConnections
from mininet.log import setLogLevel, info
from mininet.link import TCLink
from mininet.util import pmonitor

class CustomTopo(Topo):
	"Single switch connected to n hosts."
	def build(self, net_topo):
		
		for net in net_topo['networks'].keys():
			self.addSwitch(net)
			for net_ in net_topo['networks'][net].keys():
				args = net_topo['networks'][net][net_]
				self.addLink(net, net_,**args)
		
		for peer in net_topo['peers'].keys():
			self.addHost(peer)
			for net_ in net_topo['peers'][peer].keys():
				args = net_topo['peers'][peer][net_]
				self.addLink(peer, net_, max_queue_size=10000)

def createStarTopo(nHosts):
	switch = "s0"
	
	peers = {}
	output = {}
	output["networks"] = {switch: {}}
	
	for i in range(nHosts):
		peers["h"+str(i)] = {switch: {}}

	output["peers"] = peers
	
	return output

def connectionTest(nHosts):

	call(["mn","-c"])
	
	fullTopo = createStarTopo(nHosts)

	"Create and test a simple network"
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
		Usage: sudo python3 launch.py inputPath configPath
	"""
	# Tell mininet to print useful information
	setLogLevel('info')
	connectionTest(int(sys.argv[1]))
