package utils

import "unicode"

var (
	CommandsMessage = "Поддерживаемые команды:\n/tasks - Список всех активных задач\n" +
		"/new XXX YYY ZZZ - Создать новую задачу\n/assign_$ID - Назначить задачу на себя\n" +
		"/unassign_$ID - Снять задачу с текущего исполнителя\n" +
		"/resolve_$ID - Пометить задачу как выполненную, удалить из списка\n" +
		"/my - Получить задачи, назначенные на Вас\n/owner - Получить задачи, созданные Вами"
)

func CapitalizeFirstLetter(s string) string {
	text := []rune(s)
	text[0] = unicode.ToUpper(text[0])
	return string(text)
}
