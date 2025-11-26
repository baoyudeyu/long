package dragon

import (
	"dragon-alert-bot/lottery"
	"fmt"
	"strings"
)

// PatternResult 模式检测结果
type PatternResult struct {
	PatternType   string // a, ab, abb, ab_ac, ab_cd
	AttributeType string // size, parity, sum, size_parity
	Count         int
	StartQihao    string
	CurrentQihao  string
	PatternDetail string
	Matched       bool
}

// CheckPatternA 检测 a 格式（连续相同）
// 注意：attrs 应该是从旧到新排列，检测从最新期（末尾）开始往前看
func CheckPatternA(attrs []lottery.Attributes, attrType string, minCount int) *PatternResult {
	if len(attrs) < minCount {
		return &PatternResult{Matched: false}
	}

	getValue := func(attr lottery.Attributes) string {
		switch attrType {
		case "size":
			return attr.Size
		case "parity":
			return attr.Parity
		case "sum":
			return fmt.Sprintf("%d", attr.SumValue)
		default:
			return ""
		}
	}

	// 从最新的开始检测（数组末尾）
	lastIdx := len(attrs) - 1
	count := 1
	firstValue := getValue(attrs[lastIdx])
	details := []string{firstValue}

	// 往前检测
	for i := lastIdx - 1; i >= 0; i-- {
		currentValue := getValue(attrs[i])
		if currentValue == firstValue {
			count++
			details = append([]string{currentValue}, details...) // 前插
		} else {
			break
		}
	}

	if count >= minCount {
		return &PatternResult{
			PatternType:   "a",
			AttributeType: attrType,
			Count:         count,
			StartQihao:    attrs[lastIdx-count+1].Qihao, // 起始期号
			CurrentQihao:  attrs[lastIdx].Qihao,         // 当前期号
			PatternDetail: strings.Join(details, " "),
			Matched:       true,
		}
	}

	return &PatternResult{Matched: false}
}

// CheckPatternAB 检测 ab 格式（交替出现）
// 注意：attrs 应该是从旧到新排列，检测从最新期（末尾）开始往前看
func CheckPatternAB(attrs []lottery.Attributes, attrType string, minCount int) *PatternResult {
	if len(attrs) < minCount || minCount < 2 {
		return &PatternResult{Matched: false}
	}

	getValue := func(attr lottery.Attributes) string {
		switch attrType {
		case "size":
			return attr.Size
		case "parity":
			return attr.Parity
		case "sum":
			return fmt.Sprintf("%d", attr.SumValue)
		default:
			return ""
		}
	}

	lastIdx := len(attrs) - 1
	valueA := getValue(attrs[lastIdx])   // 最新期
	valueB := getValue(attrs[lastIdx-1]) // 上一期

	if valueA == valueB {
		return &PatternResult{Matched: false}
	}

	count := 2
	details := []string{valueB, valueA}

	// 继续往前检测
	for i := lastIdx - 2; i >= 0; i-- {
		pos := lastIdx - i
		expectedValue := valueA
		if pos%2 == 1 {
			expectedValue = valueB
		}

		currentValue := getValue(attrs[i])
		if currentValue == expectedValue {
			count++
			details = append([]string{currentValue}, details...)
		} else {
			break
		}
	}

	if count >= minCount {
		return &PatternResult{
			PatternType:   "ab",
			AttributeType: attrType,
			Count:         count,
			StartQihao:    attrs[lastIdx-count+1].Qihao,
			CurrentQihao:  attrs[lastIdx].Qihao,
			PatternDetail: strings.Join(details, " "),
			Matched:       true,
		}
	}

	return &PatternResult{Matched: false}
}

// CheckPatternABB 检测 abb 格式（A-B-B模式）
// 注意：attrs 应该是从旧到新排列，检测从最新期（末尾）开始往前看
func CheckPatternABB(attrs []lottery.Attributes, attrType string, minCount int) *PatternResult {
	// abb 最少需要3个元素才能形成完整模式
	if len(attrs) < 3 || minCount < 3 {
		return &PatternResult{Matched: false}
	}

	getValue := func(attr lottery.Attributes) string {
		switch attrType {
		case "size":
			return attr.Size
		case "parity":
			return attr.Parity
		case "sum":
			return fmt.Sprintf("%d", attr.SumValue)
		default:
			return ""
		}
	}

	lastIdx := len(attrs) - 1

	// 从最新期往前倒推，检测是否符合 A-B-B 模式
	// 最新3期应该是: A-B-B
	if lastIdx >= 2 {
		val0 := getValue(attrs[lastIdx])
		val1 := getValue(attrs[lastIdx-1])
		val2 := getValue(attrs[lastIdx-2])

		// 检查是否是 A-B-B
		if val2 != val1 && val1 == val0 {
			patternA := val2
			patternB := val1
			count := 3
			details := []string{patternA, patternB, patternB}

			// 继续往前检测，按 A-B-B 循环
			i := lastIdx - 3
			for i >= 0 {
				// 接下来应该是 A-B-B 的循环
				if i >= 2 {
					nextA := getValue(attrs[i])
					nextB1 := getValue(attrs[i+1])
					nextB2 := getValue(attrs[i+2])

					if nextA == patternA && nextB1 == patternB && nextB2 == patternB {
						details = append([]string{nextA, nextB1, nextB2}, details...)
						count += 3
						i -= 3
					} else {
						break
					}
				} else {
					break
				}
			}

			// 只保留完整的abb组（3的倍数）
			completeGroups := (count / 3) * 3
			if completeGroups >= minCount {
				startIdx := lastIdx - completeGroups + 1
				return &PatternResult{
					PatternType:   "abb",
					AttributeType: attrType,
					Count:         completeGroups,
					StartQihao:    attrs[startIdx].Qihao,
					CurrentQihao:  attrs[lastIdx].Qihao,
					PatternDetail: strings.Join(details, " "),
					Matched:       true,
				}
			}
		}
	}

	return &PatternResult{Matched: false}
}

// CheckPatternABAC 检测 ab,ac 格式（第一属性固定，第二属性交替）
// 注意：attrs 应该是从旧到新排列
func CheckPatternABAC(attrs []lottery.Attributes, minCount int) *PatternResult {
	if len(attrs) < minCount || minCount < 2 {
		return &PatternResult{Matched: false}
	}

	lastIdx := len(attrs) - 1
	// 第一属性固定，第二属性交替
	firstAttr1 := attrs[lastIdx].Size
	secondAttr1 := attrs[lastIdx].Parity

	firstAttr2 := attrs[lastIdx-1].Size
	secondAttr2 := attrs[lastIdx-1].Parity

	// 第一属性必须相同，第二属性必须不同
	if firstAttr1 != firstAttr2 || secondAttr1 == secondAttr2 {
		return &PatternResult{Matched: false}
	}

	count := 2
	details := []string{
		fmt.Sprintf("%s%s", firstAttr2, secondAttr2),
		fmt.Sprintf("%s%s", firstAttr1, secondAttr1),
	}

	for i := lastIdx - 2; i >= 0; i-- {
		pos := lastIdx - i
		expectedSecond := secondAttr1
		if pos%2 == 1 {
			expectedSecond = secondAttr2
		}

		if attrs[i].Size == firstAttr1 && attrs[i].Parity == expectedSecond {
			count++
			details = append([]string{fmt.Sprintf("%s%s", attrs[i].Size, attrs[i].Parity)}, details...)
		} else {
			break
		}
	}

	if count >= minCount {
		return &PatternResult{
			PatternType:   "ab_ac",
			AttributeType: "size_parity",
			Count:         count,
			StartQihao:    attrs[lastIdx-count+1].Qihao,
			CurrentQihao:  attrs[lastIdx].Qihao,
			PatternDetail: strings.Join(details, " "),
			Matched:       true,
		}
	}

	return &PatternResult{Matched: false}
}

// CheckPatternABCD 检测 ab,cd 格式（两个属性同时交替）
// 注意：attrs 应该是从旧到新排列
func CheckPatternABCD(attrs []lottery.Attributes, minCount int) *PatternResult {
	if len(attrs) < minCount || minCount < 2 {
		return &PatternResult{Matched: false}
	}

	lastIdx := len(attrs) - 1
	firstAttr1 := attrs[lastIdx].Size
	secondAttr1 := attrs[lastIdx].Parity

	firstAttr2 := attrs[lastIdx-1].Size
	secondAttr2 := attrs[lastIdx-1].Parity

	// 两个属性都必须不同
	if firstAttr1 == firstAttr2 || secondAttr1 == secondAttr2 {
		return &PatternResult{Matched: false}
	}

	count := 2
	details := []string{
		fmt.Sprintf("%s%s", firstAttr2, secondAttr2),
		fmt.Sprintf("%s%s", firstAttr1, secondAttr1),
	}

	for i := lastIdx - 2; i >= 0; i-- {
		pos := lastIdx - i
		expectedFirst := firstAttr1
		expectedSecond := secondAttr1
		if pos%2 == 1 {
			expectedFirst = firstAttr2
			expectedSecond = secondAttr2
		}

		if attrs[i].Size == expectedFirst && attrs[i].Parity == expectedSecond {
			count++
			details = append([]string{fmt.Sprintf("%s%s", attrs[i].Size, attrs[i].Parity)}, details...)
		} else {
			break
		}
	}

	if count >= minCount {
		return &PatternResult{
			PatternType:   "ab_cd",
			AttributeType: "size_parity",
			Count:         count,
			StartQihao:    attrs[lastIdx-count+1].Qihao,
			CurrentQihao:  attrs[lastIdx].Qihao,
			PatternDetail: strings.Join(details, " "),
			Matched:       true,
		}
	}

	return &PatternResult{Matched: false}
}

// CheckPatternABAB 检测 abab 格式（组合重复：小单小单 或 大双大双）
// 注意：attrs 应该是从旧到新排列
func CheckPatternABAB(attrs []lottery.Attributes, minCount int) *PatternResult {
	if len(attrs) < minCount || minCount < 2 {
		return &PatternResult{Matched: false}
	}

	lastIdx := len(attrs) - 1
	// 获取最新的组合
	combo1 := fmt.Sprintf("%s%s", attrs[lastIdx].Size, attrs[lastIdx].Parity)

	count := 1
	details := []string{combo1}

	// 往前检查是否都是相同组合
	for i := lastIdx - 1; i >= 0; i-- {
		currentCombo := fmt.Sprintf("%s%s", attrs[i].Size, attrs[i].Parity)
		if currentCombo == combo1 {
			count++
			details = append([]string{currentCombo}, details...)
		} else {
			break
		}
	}

	if count >= minCount {
		return &PatternResult{
			PatternType:   "abab",
			AttributeType: "size_parity",
			Count:         count,
			StartQihao:    attrs[lastIdx-count+1].Qihao,
			CurrentQihao:  attrs[lastIdx].Qihao,
			PatternDetail: strings.Join(details, " "),
			Matched:       true,
		}
	}

	return &PatternResult{Matched: false}
}
