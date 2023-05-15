#!/usr/bin/env python
# coding: utf-8

from math import ceil

def createTreeTopo(nHosts, max_children=2,
		hostBand=10, hostQueue=1000, hostDelay="5ms", hostLoss=0):

	"""
		The topology is distributed like a tree.
		Each node will contain up to 19 childrens and 1 parent.
		Each leaf will have it's children as hosts
	"""
	
	tree = createTree(nHosts, n_children=max_children)
	edges = makeLinks(tree)

	network = {}
	peers = {}
	output = {}

	done = set()
	for i in range(1,len(tree)):
		layer = tree[i]
		for switch1 in layer:
			links = {}
			for switch2 in done:
				if {switch1,switch2} in edges:
					links[switch2] = {}
			network[switch1] = links
			done.add(switch1)

	output["networks"] = network

	for host in tree[0]:
		for switch in tree[1]:
			if {host,switch} in edges:
				peers[host] = {switch: {
					"bw": hostBand,
					"delay": hostDelay,
					"loss": hostLoss,
					"max_queue_size": hostQueue
				}}
				break
				
	# Adding main server
	peers["h0"] = {tree[len(tree)-1][-1]: {}}

	output["peers"] = peers

	return output
	
def createTree(nHosts, n_children):
	""" Creates the tree structure """
	layer = 0
	tree = {0: ["h"+str(i+1) for i in range(nHosts)]}
	nSwitches = 0
	while len(tree[layer]) > 1:
		nNodes = ceil(len(tree[layer])/n_children)
		switches = ["s"+str(i) for i in range(nSwitches,nSwitches+nNodes)]
		layer += 1
		nSwitches += nNodes
		tree[layer] = switches
		
	return tree
	
def makeLinks(tree):
	""" Create links between nodes according to tree structure """
	links = set()
	for i in range(1,len(tree)):
		for link in makeLinksLayer(tree[i-1],tree[i]):
			links.add(link)

	return links
	
def makeLinksLayer(children, parent):
	""" Create node links between layers """
	nC = len(children)
	nP = len(parent)
	nLinks = ceil(nC/nP)

	links = set()
	c = 0
	for p in parent:
		for _ in range(nLinks):
			links.add(frozenset([p,children[c]]))
			c += 1
			if c == nC:
				break
	return links
