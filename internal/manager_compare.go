package internal

const (
	ActionUpload = "upload"
	ActionDelete = "delete"
	ActionChange = "change"
)

type CompareManager struct {
}

func NewCompareManager() *CompareManager {
	return &CompareManager{}
}

func (m *CompareManager) Compare(source map[string]CompareObject, remote map[string]CompareObject) []CompareResult {
	result := make([]CompareResult, 0, 100)
	for _, localObject := range source {
		remoteObject, ok := remote[localObject.Path()]
		if !ok {
			result = append(result, CompareResult{Object: localObject, Action: ActionUpload})
		} else if localObject.Hash() != remoteObject.Hash() {
			result = append(result, CompareResult{Object: localObject, Action: ActionChange})
		}
	}
	for _, remoteObject := range remote {
		_, ok := source[remoteObject.Path()]
		if !ok {
			result = append(result, CompareResult{Object: remoteObject, Action: ActionDelete})
		}
	}
	return result
}
