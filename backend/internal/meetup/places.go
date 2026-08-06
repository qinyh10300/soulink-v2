package meetup

var campusPlaces = []CampusPlace{
	{Name: "紫荆园一层", Area: "紫荆生活区", BestFor: []string{"meal", "conversation"}},
	{Name: "清芬园", Area: "学堂路", BestFor: []string{"meal", "conversation"}},
	{Name: "李文正图书馆公共区", Area: "北馆", BestFor: []string{"study", "project"}},
	{Name: "学生服务中心开放区", Area: "C 楼", BestFor: []string{"project", "conversation"}},
	{Name: "西大操场入口", Area: "西区", BestFor: []string{"sport", "event"}},
	{Name: "东大操场看台", Area: "东区", BestFor: []string{"sport", "conversation"}},
	{Name: "情人坡主路口", Area: "近大礼堂", BestFor: []string{"romantic", "conversation"}},
}

func placesFor(category string) []CampusPlace {
	preferred := make([]CampusPlace, 0, 4)
	rest := make([]CampusPlace, 0, len(campusPlaces))
	for _, place := range campusPlaces {
		matched := false
		for _, bestFor := range place.BestFor {
			if bestFor == category {
				matched = true
				break
			}
		}
		if matched {
			preferred = append(preferred, place)
		} else {
			rest = append(rest, place)
		}
	}
	preferred = append(preferred, rest...)
	if len(preferred) > 5 {
		preferred = preferred[:5]
	}
	return preferred
}
