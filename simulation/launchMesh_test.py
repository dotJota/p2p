#!/usr/bin/python                                                                            
     
from time import time                   
from time import sleep        
from signal import SIGINT
from subprocess import call
import json
import sys
import os
                                                                                             
from mininet.topo import Topo
from mininet.net import Mininet
from mininet.util import dumpNodeConnections
from mininet.log import setLogLevel, info
from mininet.node import CPULimitedHost, Controller, RemoteController
from mininet.link import TCULink
from mininet.util import pmonitor

class CustomTopo(Topo):
	"Single switch connected to n hosts."
	def build(self, net_topo):
		
		for net in net_topo['networks'].keys():
			self.addSwitch(net)
			if net_topo['networks'][net] != None:
				for net_ in net_topo['networks'][net].keys():
					args = net_topo['networks'][net][net_]
					self.addLink(net, net_,**args)
		
		n = len(net_topo['peers'].keys())
		for peer in net_topo['peers'].keys():
			self.addHost(peer)
			for net_ in net_topo['peers'][peer].keys():
				args = net_topo['peers'][peer][net_]
				self.addLink(peer, net_,**args)

def meshRules(nHosts, nSwitches):
	hSwitches = round(nHosts**(1/2)+0.5)
	
	sRanges = []
	for i in range(nSwitches):
		sRange = [k for k in range(hSwitches*i,hSwitches*(i+1))]
		sRanges.append(sRange)
		
	for i in range(nSwitches):
		switch = "s" + str(i)
		orderDict = orderSwitches(i, nSwitches)
		for j in range(nHosts):
			target = findSwitch(j,sRanges)
			if target == i:
				offset1 = nSwitches - 1
				offset2 = j - hSwitches*i
				position = offset1 + offset2 + 1
				cmd = f'ovs-ofctl add-flow {switch} table=0,idle_timeout=60,priority=100,dl_type=0x0800,nw_dst=10.0.0.{j+1},actions=output:"{switch}-eth{position}"'
				print(cmd)
				os.system(cmd)
				cmd = f'ovs-ofctl add-flow {switch} table=0,idle_timeout=60,priority=100,dl_type=0x0806,nw_dst=10.0.0.{j+1},actions=output:"{switch}-eth{position}"'
				print(cmd)
				os.system(cmd)
			else:
				position = orderDict[target]
				cmd = f'ovs-ofctl add-flow {switch} table=0,idle_timeout=60,priority=100,dl_type=0x0800,nw_dst=10.0.0.{j+1},actions=output:"{switch}-eth{position}"'
				print(cmd)
				os.system(cmd)
				cmd = f'ovs-ofctl add-flow {switch} table=0,idle_timeout=60,priority=100,dl_type=0x0806,nw_dst=10.0.0.{j+1},actions=output:"{switch}-eth{position}"'
				print(cmd)
				os.system(cmd)
				

def findSwitch(node, sRanges):
	i = 0
	for _ in range(len(sRanges)):
		if node in sRanges[i]:
			break
		i+=1
	return i
			
def orderSwitches(s, nSwitches):
	orderDict = {}
	counter = 1
	for i in range(nSwitches):
		if i != s:
			orderDict[i] = counter
			counter += 1
	return orderDict

def meshTopo(nHosts, hostBand=10, hostQueue=1000, hostDelay=0, hostLoss=0):
	"Fully connected switches connected to n/s hosts."
	hSwitches = round(nHosts**(1/2)+0.5)
	
	links = set()
	i = 0
	j = 0
	while i < nHosts:
		s = "s"+str(j)
		while i < nHosts:
			h = "h"+str(i)
			link = frozenset([h,s])
			links.add(link)
			i+=1
			if (i%hSwitches)==0:
				break
		j+=1
	
	nSwitches = j
	switches = []
	for s in range(nSwitches):
		switches.append("s"+str(s))
	
	hosts = ["h"+str(i) for i in range(nHosts)]
	hostDict = distributeHosts(hosts, switches, links, hostBand, hostQueue, hostDelay, hostLoss)

	links = set()
	for s1 in switches:
		for s2 in switches:
			if s1 != s2:
				links.add(frozenset([s1,s2]))
					
	netDict = createDictSwitches(switches,links)
	
	outDict = {}
	outDict["networks"] = netDict
	outDict["peers"] = hostDict
	
	return outDict, nSwitches
	

def createDictSwitches(switches, links):
    
	''' Creates a the network topo defined by edges'''

	netOutput = {}

	done = set()
	for s1 in switches:
		done.add(s1)

		edges = {}
		for s2 in done:
			if getParam(s1,s2,links):
				edges[s2] = {"loss":0}

		netOutput[s1] = None

		if len(edges) != 0:
			netOutput[s1] = edges

	return netOutput

def distributeHosts(hosts, switches, links, hostBand, hostQueue, hostDelay, hostLoss):
	hostOutput = {}

	for h in hosts:
		for s in switches:
			if {h,s} in links:
				hostOutput[h] = {s: {
					"bw": hostBand,
					"delay": hostDelay,
					"loss": hostLoss,
					"max_queue_size": hostQueue
				}}

	return hostOutput

def getParam(s1, s2, links):
	if {s1,s2} in links:
		return True
	return False

		
def simpleTest(nHosts):

	call(["mn","-c"])
	
	fullTopo, nSwitches = meshTopo(nHosts)

	topo = CustomTopo(fullTopo)
	
	net = Mininet(topo, link=TCULink, build=False, waitConnected=False)
	net.addController( RemoteController( 'c0', ip='127.0.0.1', port=6633 ))
	net.build()
	
	try:
		net.start()
		
		print( "Dumping switch connections" )
		dumpNodeConnections(net.switches)

		meshRules(nHosts, nSwitches)

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
	simpleTest(int(sys.argv[1]))
