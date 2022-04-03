package url

//GetScheme возвращает правильный формат протокола в зависимости от
//того, безопасное ли соединение. На данный момент поддерживает только
//протоколы "ws" и "http".
func GetScheme(protocol string, secure bool) string {
	switch protocol {
	case "ws":
		switch secure {
		case true:
			return "wss://"
		case false:
			return "ws://"
		}
	case "http":
		switch secure {
		case true:
			return "https://"
		case false:
			return "http://"
		}
	}
	return ""
}
