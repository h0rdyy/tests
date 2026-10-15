package content

type Service struct {
	ID                       int
	Title, Hint, Cell, Thumb string
}

var Services = []Service{
	{1, "Защита персональных данных", "Ваш юридический щит и репутационный актив", "col1", "shield.png"},
	{2, "Услуги по обеспечению информационной безопасности", "Технологическая крепость вашего бизнеса", "top", "fire.png"},
	{3, "Информационное право", "", "bl", "lightning.png"},
	{4, "Разработка образовательных программ", "", "br", "light.png"},
	{5, "Обучение и развитие персонала", "Инвестиции в вашу безопасность и репутацию", "col4", "study.png"},
}

func FindService(id int) (Service, bool) {
	for _, service := range Services {
		if service.ID == id {
			return service, true
		}
	}
	return Service{}, false
}
