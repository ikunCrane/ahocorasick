package goahocorasick

import (
	"fmt"
)

import (
	"github.com/anknown/darts"
)

const FAIL_STATE = -1
const ROOT_STATE = 1

// 新增结构体：将模式串与自定义数据绑定
type PatternData struct {
	Pattern []rune      // 原始模式串
	Data    interface{} // 自定义数据（可以是任意类型）
}

type Machine struct {
	trie    *godarts.DoubleArrayTrie
	failure []int
	output  map[int][]PatternData // 每个状态对应多个PatternData
}

type Term struct {
	Pos        int
	Word       []rune
	CustomData interface{}
}

func (m *Machine) Build(keywords [][]rune) (err error) {
	return m.BuildByCustom(keywords, nil)
}

func (m *Machine) BuildByCustom(keywords [][]rune, customdata map[string]interface{}) (err error) {
	if len(keywords) == 0 {
		return fmt.Errorf("empty keywords")
	}

	d := new(godarts.Darts)

	trie := new(godarts.LinkedListTrie)
	m.trie, trie, err = d.Build(keywords)
	if err != nil {
		return err
	}
	m.output = make(map[int][]PatternData, 0)

	for idx, val := range d.Output {

		pattern := new(PatternData)
		if _, ok := customdata[string(val)]; ok != false {
			pattern.Data = customdata[string(val)]
		}
		pattern.Pattern = val
		m.output[idx] = append(m.output[idx], *pattern)
	}

	queue := make([](*godarts.LinkedListTrieNode), 0)
	m.failure = make([]int, len(m.trie.Base))
	for _, c := range trie.Root.Children {
		m.failure[c.Base] = godarts.ROOT_NODE_BASE
	}
	queue = append(queue, trie.Root.Children...)

	for {
		if len(queue) == 0 {
			break
		}

		node := queue[0]
		for _, n := range node.Children {
			if n.Base == godarts.END_NODE_BASE {
				continue
			}
			inState := m.f(node.Base)
		set_state:
			outState := m.g(inState, n.Code-godarts.ROOT_NODE_BASE)
			if outState == FAIL_STATE {
				inState = m.f(inState)
				goto set_state
			}
			if _, ok := m.output[outState]; ok != false {
				/*copyOutState := make([][]rune, 0)
				for _, o := range m.output[outState] {
					copyOutState = append(copyOutState, o)
				}
				m.output[n.Base] = append(copyOutState, m.output[n.Base]...)
				*/
				copyOutState := make([]PatternData, 0)
				for _, o := range m.output[outState] {
					copyOutState = append(copyOutState, o)
				}
				m.output[n.Base] = append(copyOutState, m.output[n.Base]...)
			}
			m.setF(n.Base, outState)
		}
		queue = append(queue, node.Children...)
		queue = queue[1:]
	}

	return nil
}
func (m *Machine) PrintFailure() {
	fmt.Printf("+-----+-----+\n")
	fmt.Printf("|%5s|%5s|\n", "index", "value")
	fmt.Printf("+-----+-----+\n")
	for i, v := range m.failure {
		fmt.Printf("|%5d|%5d|\n", i, v)
	}
	fmt.Printf("+-----+-----+\n")
}

func (m *Machine) PrintOutput() {
	fmt.Printf("+-----+----------+\n")
	fmt.Printf("|%5s|%10s|\n", "index", "value")
	fmt.Printf("+-----+----------+\n")
	for i, v := range m.output {
		var val string
		for _, o := range v {
			val = val + " " + string(o.Pattern)
		}
		fmt.Printf("|%5d|%10s|\n", i, val)
	}
	fmt.Printf("+-----+----------+\n")
}

func (m *Machine) g(inState int, input rune) (outState int) {
	if inState == FAIL_STATE {
		return ROOT_STATE
	}

	t := inState + int(input) + godarts.ROOT_NODE_BASE
	if t >= len(m.trie.Base) {
		if inState == ROOT_STATE {
			return ROOT_STATE
		}
		return FAIL_STATE
	}
	if inState == m.trie.Check[t] {
		return m.trie.Base[t]
	}

	if inState == ROOT_STATE {
		return ROOT_STATE
	}

	return FAIL_STATE
}

func (m *Machine) f(index int) (state int) {
	return m.failure[index]
}

func (m *Machine) setF(inState, outState int) {
	m.failure[inState] = outState
}

func (m *Machine) MultiPatternSearch(content []rune, returnImmediately bool) [](*Term) {
	terms := make([](*Term), 0)

	state := ROOT_STATE
	for pos, c := range content {
	start:
		if m.g(state, c) == FAIL_STATE {
			state = m.f(state)
			goto start
		} else {
			state = m.g(state, c)
			if val, ok := m.output[state]; ok != false {
				for _, word := range val {
					term := new(Term)
					term.Pos = pos - len(word.Pattern) + 1
					term.Word = word.Pattern
					term.CustomData = word.Data
					terms = append(terms, term)
					if returnImmediately {
						return terms
					}
				}
			}
		}
	}

	return terms
}

func (m *Machine) ExactSearch(content []rune) [](*Term) {
	if m.trie.ExactMatchSearch(content, 0) {
		t := new(Term)
		t.Word = content
		t.Pos = 0
		return [](*Term){t}
	}

	return nil
}
