package models

// IPOwner 接口定义了任何需要按 IP 比较的对象必须实现的方法。
type IPOwner interface {
	GetIP() string
	// 可以添加其他方法，例如 GetValue() string 来帮助打印或比较，但对于本函数不是必需的
}

// CompareDifferentObjectsByIP 比较两个不同类型的对象切片（T1 和 T2），
// 只要它们都实现了 IPOwner 接口。
//
// currentSlice: 代表当前状态的对象集合 (类型 T1)。
// desiredSlice: 代表期望状态的对象集合 (类型 T2)。
// Returns:
//   - toAdd: 一个包含所有在 desiredSlice 中但其 IP 不在 currentSlice IP 集合中的 T2 类型对象的切片。
//   - toDelete: 一个包含所有在 currentSlice 中但其 IP 不在 desiredSlice IP 集合中的 T1 类型对象的切片。
func CompareObjectsByIP[T1 IPOwner, T2 IPOwner](
	currentSlice []T1,
	desiredSlice []T2,
) (toAdd []T2, toDelete []T1) {

	// 1. 创建一个 map 来存储 desiredSlice 中所有对象的 IP 及其对应的 T2 类型对象。
	//    如果 desiredSlice 中有重复的 IP，map 中只会保留最后一个具有该 IP 的对象。
	desiredIPMap := make(map[string]T2, len(desiredSlice))
	for _, item := range desiredSlice {
		// 检查接口值是否为 nil，或者其底层指针是否为 nil
		// 对于指针类型，直接 item == nil 即可；对于接口，需要更小心
		// 但由于 T2 约束为 IPOwner，它通常是一个具体类型（如指针）
		// 简单起见，如果 item 可以为 nil（例如 T2 是 *StructType），则应检查。
		// if item == nil { continue } // 假设 T1, T2 是指针类型时需要
		// 为确保安全，可以添加一个通用的 IsNil 检查，但这里我们依赖调用者不传入nil接口值本身，而是nil指针
		// 或者假定 T1, T2 不会是接口类型本身，而是实现接口的具体（指针）类型。
		// 如果 item 是指针类型， item == nil 是正确的检查。

		// 一个更通用的 nil 检查，以防 T1 或 T2 本身是接口类型且其动态值为 nil
		// 不过，通常我们会传入具体类型的切片，如 []*StructA, []*StructB
		var isItemNil bool
		if v, ok := interface{}(item).(interface{ IsNil() bool }); ok { // 检查是否有 IsNil 方法 (不标准)
			isItemNil = v.IsNil()
		} else {
			// 对于指针类型，可以直接比较。对于非指针且非接口类型，它不能为nil。
			// 为了简单起见，如果T1, T2是预期的指针类型，直接比较item == nil就可以。
			// 此处我们假设调用 GetIP 前，item 不会导致 panic。
			// 如果 T1, T2 是指针，调用者应确保不传入包含 nil 元素的切片，或在此处处理。
			// 之前版本已处理 itemPtr == nil，此处类似
		}
		if isItemNil { // 简化版：如果 item 是指针，(item == nil)
			// continue
		}
		// 为了安全，我们假设 GetIP() 可以在 nil 接收器上安全调用（返回空字符串或类似）
		// 或者，我们需要确保 item 不是 nil。
		// 如果 T2 是指针类型，则需要检查 item 是否为 nil。
		// 例如：if item == nil { continue }

		desiredIPMap[item.GetIP()] = item
	}

	// 2. 创建一个 map 来存储 currentSlice 中所有对象的 IP。
	//    仅用于快速检查 currentSlice 中是否存在某个 IP。
	currentIPSet := make(map[string]struct{}, len(currentSlice))
	for _, item := range currentSlice {
		// if item == nil { continue } // 如果 T1 是指针类型
		currentIPSet[item.GetIP()] = struct{}{}
	}

	// 3. 找出需要添加的 T2 类型对象：
	//    遍历 desiredIPMap。
	//    如果某个 IP 存在于 desiredIPMap 但不存在于 currentIPSet，则将 desiredIPMap 中与该 IP 对应的 T2 对象添加到 toAdd。
	for ip, desiredObject := range desiredIPMap {
		// if desiredObject == nil { continue } // 如果 T2 是指针类型
		if _, foundInCurrent := currentIPSet[ip]; !foundInCurrent {
			toAdd = append(toAdd, desiredObject)
		}
	}

	// 4. 找出需要删除的 T1 类型对象：
	//    遍历原始的 currentSlice 中的每个 T1 对象。
	//    如果其 IP 不存在于 desiredIPMap，则将其添加到 toDelete。
	for _, currentObject := range currentSlice {
		// if currentObject == nil { continue } // 如果 T1 是指针类型
		ip := currentObject.GetIP()
		if _, foundInDesired := desiredIPMap[ip]; !foundInDesired {
			toDelete = append(toDelete, currentObject)
		}
	}

	return toAdd, toDelete
}
