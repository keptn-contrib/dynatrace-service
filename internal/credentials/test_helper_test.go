package credentials

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testDynatraceAPIToken = "dt0c01.ST2EY72KQINMH574WMNVI7YN.G3DFPBEJYMODIDAEX454M7YWBUVEFOWKPRVMWFASS64NFH52PX6BNDVFFM572RZM"

func createTestSecret(name string, data map[string]string) *v1.Secret {
	convertedData := make(map[string][]byte)
	for key, value := range data {
		convertedData[key] = []byte(value)
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "keptn",
		},
		Data: convertedData,
	}
}
