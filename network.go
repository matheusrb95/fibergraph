package main

func network1() *Node {
	Box1 := NewNode(ID(), "Box1", BoxNodeType)
	Box2 := NewNode(ID(), "Box2", BoxNodeType)
	Box3 := NewNode(ID(), "Box3", BoxNodeType)

	ONU1 := NewNode(ID(), "ONU1", ONUNodeType)
	ONU1.Active = false

	ONU2 := NewNode(ID(), "ONU2", ONUNodeType)
	ONU2.Active = false

	ONU3 := NewNode(ID(), "ONU3", ONUNodeType)
	ONU3.Active = false

	ONU4 := NewNode(ID(), "ONU4", ONUNodeType)
	//ONU4.Active = false

	OLT := NewNode(ID(), "OLT", OLTNodeType)

	Splitter1 := NewNode(ID(), "Splitter1", SplitterNodeType)
	Splitter2 := NewNode(ID(), "Splitter2", SplitterNodeType)
	Splitter3 := NewNode(ID(), "Splitter3", SplitterNodeType)

	Box1.SetParent(OLT)
	Box1.SetChildren(Splitter1)
	Splitter1.SetChildren(ONU1, ONU2)

	Box2.SetParent(OLT)
	Box2.SetChildren(Splitter2)
	Splitter2.SetChildren(ONU3)

	Box3.SetParent(OLT)
	Box3.SetChildren(Splitter3)
	Splitter3.SetChildren(ONU4)

	ONU1.SetParent(Splitter1)
	ONU2.SetParent(Splitter1)
	ONU3.SetParent(Splitter2)
	ONU4.SetParent(Splitter3)

	OLT.SetChildren(Box1, Box2, Box3)

	return OLT
}

func network2() *Node {
	OLT := NewNode(ID(), "OLT", OLTNodeType)

	Cable1 := NewNode(ID(), "Cable1", CableNodeType)

	CEO1 := NewNode(ID(), "CEO1", BoxNodeType)
	Splitter1 := NewNode(ID(), "Splitter1x2", SplitterNodeType)

	Cable2 := NewNode(ID(), "Cable2", CableNodeType)
	Cable2.Active = true

	CTO1 := NewNode(ID(), "CTO1", BoxNodeType)
	Splitter2 := NewNode(ID(), "Splitter1x4", SplitterNodeType)

	Cable3 := NewNode(ID(), "Cable3", CableNodeType)

	CTO2 := NewNode(ID(), "CTO2", BoxNodeType)
	Splitter3 := NewNode(ID(), "Splitter1x4", SplitterNodeType)

	Drop1 := NewNode(ID(), "Drop1", CableNodeType)
	Drop2 := NewNode(ID(), "Drop2", CableNodeType)
	Drop3 := NewNode(ID(), "Drop3", CableNodeType)
	Drop4 := NewNode(ID(), "Drop4", CableNodeType)
	Drop5 := NewNode(ID(), "Drop5", CableNodeType)

	ONU1 := NewNode(ID(), "ONU1", ONUNodeType)
	ONU1.Active = false
	ONU2 := NewNode(ID(), "ONU2", ONUNodeType)
	ONU2.Active = true
	ONU3 := NewNode(ID(), "ONU3", ONUNodeType)
	ONU3.Active = false
	ONU4 := NewNode(ID(), "ONU4", ONUNodeType)
	ONU4.Active = false
	ONU5 := NewNode(ID(), "ONU5", ONUNodeType)
	ONU5.Active = false

	OLT.SetChildren(Cable1)

	Cable1.SetParent(OLT)
	Cable1.SetChildren(CEO1)

	CEO1.SetParent(Cable1)
	CEO1.SetChildren(Splitter1)

	Splitter1.SetParent(CEO1)
	Splitter1.SetChildren(Cable2, Cable3)

	Cable2.SetParent(Splitter1)
	Cable2.SetChildren(CTO1)
	Cable3.SetParent(Splitter1)
	Cable3.SetChildren(CTO2)

	CTO1.SetParent(Cable2)
	CTO1.SetChildren(Splitter2)
	CTO2.SetParent(Cable3)
	CTO2.SetChildren(Splitter3)

	Splitter2.SetParent(CTO1)
	Splitter2.SetChildren(Drop1, Drop2)

	Splitter3.SetParent(CTO2)
	Splitter3.SetChildren(Drop3, Drop4, Drop5)

	Drop1.SetParent(Splitter2)
	Drop1.SetChildren(ONU1)
	Drop2.SetParent(Splitter2)
	Drop2.SetChildren(ONU2)
	Drop3.SetParent(Splitter3)
	Drop3.SetChildren(ONU3)
	Drop4.SetParent(Splitter3)
	Drop4.SetChildren(ONU4)
	Drop5.SetParent(Splitter3)
	Drop5.SetChildren(ONU5)

	ONU1.SetParent(Drop1)
	ONU2.SetParent(Drop2)

	ONU3.SetParent(Drop3)
	ONU4.SetParent(Drop4)
	ONU5.SetParent(Drop5)

	return OLT
}
