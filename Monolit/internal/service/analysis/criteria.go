package analysis

type analysisCriterion struct {
	Code        string
	Title       string
	Description string
	PointsMax   int
	Order       int
}

var baseAnalysisCriteria = []analysisCriterion{
	{Code: "greeting", Title: "Приветствие", Description: "Менеджер корректно начал разговор, представился или обозначил контекст.", PointsMax: 10, Order: 1},
	{Code: "needs_discovery", Title: "Выявление потребности", Description: "Менеджер выявил потребность, задачу, ограничения клиента.", PointsMax: 10, Order: 2},
	{Code: "question_quality", Title: "Качество вопросов", Description: "Вопросы были открытыми, уточняющими и полезными.", PointsMax: 10, Order: 3},
	{Code: "answer_quality", Title: "Качество ответов", Description: "Ответы менеджера были понятными, точными и закрывали вопросы клиента.", PointsMax: 10, Order: 4},
	{Code: "solution_relevance", Title: "Релевантность решения", Description: "Предложение или объяснение было связано с потребностью клиента.", PointsMax: 10, Order: 5},
	{Code: "objection_handling", Title: "Работа с возражениями", Description: "Возражения клиента были обработаны, если они были.", PointsMax: 10, Order: 6},
	{Code: "pricing_clarity", Title: "Ясность цены и условий", Description: "Цена, условия, сроки или ограничения были объяснены, если обсуждались.", PointsMax: 10, Order: 7},
	{Code: "tone_professionalism", Title: "Профессиональный тон", Description: "Тон был вежливым, спокойным, профессиональным.", PointsMax: 10, Order: 8},
	{Code: "next_step_quality", Title: "Качество следующего шага", Description: "Следующий шаг был конкретным, с ответственным или сроком, если разговор должен продолжаться.", PointsMax: 10, Order: 9},
	{Code: "outcome_clarity", Title: "Ясность итога", Description: "Итог разговора понятен.", PointsMax: 10, Order: 10},
	{Code: "custom_instruction_match", Title: "Выполнение дополнительной инструкции", Description: "Дополнительные инструкции компании, отдела или пользователя были учтены, если они есть.", PointsMax: 10, Order: 11},
}

func analysisCriterionByCode(code string) (analysisCriterion, bool) {
	for _, criterion := range baseAnalysisCriteria {
		if criterion.Code == code {
			return criterion, true
		}
	}
	return analysisCriterion{}, false
}
