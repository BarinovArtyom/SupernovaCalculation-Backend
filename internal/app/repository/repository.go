package repository

import (
	"fmt"
	"slices"
)

type Scope struct {
	ID        int
	Name      string
	Desc      string
	Diam      string
	Filter    string
	Lambda    string
	DeltaLamb string
	ZeroPoint string
	ImgLink   string
}

type Star struct {
	ID            int
	Name          string
	Constellation string
	Status        bool
}

type Calc struct {
	ID       int
	StarID   int
	Scope_ID int
	InpVal   string
	ResVal   string
}

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

func (r *Repository) GetScopes() []Scope {
	var Scopes = []Scope{
		{
			ID:        1,
			Name:      "Hubble",
			Desc:      "Самый известный космический телескоп, работающий на орбите с 1990 года. Сделал множество революционных открытий в астрономии.",
			Diam:      "2.4",
			Filter:    "F775W.",
			Lambda:    "775",
			DeltaLamb: "150",
			ZeroPoint: "2.518E-9",
			ImgLink:   "http://127.0.0.1:9000/vedro/0.jpg",
		},
		{
			ID:        2,
			Name:      "James Webb",
			Desc:      "Космический телескоп нового поколения, преемник «Хаббла». Работает в инфракрасном диапазоне для изучения самых далёких и древних объектов Вселенной.",
			Diam:      "2.4",
			Filter:    "F775W.",
			Lambda:    "775",
			DeltaLamb: "150",
			ZeroPoint: "2.518E-9",
			ImgLink:   "http://127.0.0.1:9000/vedro/1.jpg",
		},
		{
			ID:        3,
			Name:      "W. M. Keck",
			Desc:      "Один из крупнейших оптических телескопов в мире с сегментным зеркалом. Используется для спектроскопии и получения изображений с высоким разрешением.",
			Diam:      "1150",
			Filter:    "F775W.",
			Lambda:    "775",
			DeltaLamb: "150",
			ZeroPoint: "2.518E-9",
			ImgLink:   "http://127.0.0.1:9000/vedro/2.jpg",
		},
		{
			ID:        4,
			Name:      "Gaia",
			Desc:      "Космическая обсерватория Европейского космического агентства, главная задача которой — составить сверхточную 3D-карту Млечного Пути, измерив координаты, расстояния и движения миллиардов звёзд",
			Diam:      "2.4",
			Filter:    "F775W.",
			Lambda:    "775",
			DeltaLamb: "150",
			ZeroPoint: "2.518E-9",
			ImgLink:   "http://127.0.0.1:9000/vedro/3.jpg",
		},
	}
	return Scopes
}

func (r *Repository) GetStars() []Star {
	var Stars = []Star{
		{
			ID:            1,
			Name:          "Сириус",
			Constellation: "Большой Пес",
			Status:        true,
		},
		{
			ID:            2,
			Name:          "Полярная",
			Constellation: "Малая Медведица",
			Status:        false,
		},
		{
			ID:            3,
			Name:          "Вега",
			Constellation: "Лира",
			Status:        false,
		},
		{
			ID:            4,
			Name:          "Бетельгейзе",
			Constellation: "Орион",
			Status:        false,
		},
	}
	return Stars
}

func (r *Repository) GetCalcs() []Calc {
	var Calcs = []Calc{
		{
			ID:       1,
			StarID:   1,
			Scope_ID: 1,
			InpVal:   "123",
			ResVal:   "1000",
		},
		{
			ID:       2,
			StarID:   1,
			Scope_ID: 2,
			InpVal:   "21",
			ResVal:   "10100",
		},
	}
	return Calcs
}

func (r *Repository) ScopeByID(id int, Scopes []Scope) (Scope, error) {
	i := slices.IndexFunc(Scopes, func(f Scope) bool { return f.ID == id })
	if i == -1 {
		return Scope{}, fmt.Errorf("телескоп не найден")
	}
	return Scopes[i], nil
}

func (r *Repository) StarByID(id int, Stars []Star) Star {
	return Stars[slices.IndexFunc(Stars, func(f Star) bool { return f.ID == id })]
}

func (r *Repository) CalcByID(id int, Calcs []Calc) Calc {
	return Calcs[slices.IndexFunc(Calcs, func(f Calc) bool { return f.ID == id })]
}

func (r *Repository) ScopeByStar(id int, Calcs []Calc) ([]Scope, error) {
	var MatchScope []Scope
	allScopes := r.GetScopes()

	for _, calc := range Calcs {
		if calc.StarID == id {
			scope, err := r.ScopeByID(calc.Scope_ID, allScopes)
			if err != nil {
				continue
			}
			MatchScope = append(MatchScope, scope)
		}
	}

	if len(MatchScope) == 0 {
		return nil, fmt.Errorf("телескоп не найден")
	}

	return MatchScope, nil
}

func (r *Repository) StarIDByStatus() (int, error) {
	Stars := r.GetStars()
	x := slices.IndexFunc(Stars, func(f Star) bool { return f.Status })
	if x == -1 {
		return 0, fmt.Errorf("нет активной заявки")
	}
	return Stars[x].ID, nil
}

func (r *Repository) CountCalcsByStarID(StarID int, Calcs []Calc) int {
	count := 0
	for _, calc := range Calcs {
		if calc.StarID == StarID {
			count++
		}
	}
	return count
}

func (r *Repository) CalcByStar(id int, Calcs []Calc) []Calc {
	var MatchCalc []Calc
	for _, calc := range Calcs {
		if calc.StarID == id {
			MatchCalc = append(MatchCalc, r.CalcByID(calc.ID, r.GetCalcs()))
		}
	}
	return MatchCalc
}
