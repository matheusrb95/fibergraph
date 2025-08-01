package correlation

type InconsistentCase struct {
	AlarmedSensor *Node
	ActiveSensor  *Node
}

func checkInconsistency(node *Node, result *[]*InconsistentCase) {
	for _, child := range node.Children {
		if child.AlarmedSensor() {
			if activeSensor := getActiveSensorBelow(node); activeSensor != nil {
				*result = append(*result, &InconsistentCase{
					AlarmedSensor: child,
					ActiveSensor:  activeSensor,
				})
			}
		}

		checkInconsistency(child, result)
	}
}

func getActiveSensorBelow(node *Node) *Node {
	for _, child := range node.Children {
		if child.ActiveSensor() {
			return child
		}

		if activeSensor := getActiveSensorBelow(child); activeSensor != nil {
			return activeSensor
		}
	}

	return nil
}
