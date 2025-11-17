package dbutil

import (
	"gorm.io/gorm"
)

func ApplyScopes(opts ...QueryOption) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		options := new(QueryOptions)

		for _, opt := range opts {
			opt(options)
		}

		if options.Group != "" {
			db.Group(options.Group)
		}

		if options.Joins != nil {
			for _, join := range options.Joins {
				db.Joins(join.Query, join.Args...)
			}
		}

		if options.Preloads != nil {
			for _, preload := range options.Preloads {
				db.Preload(preload.Query, preload.Args...)
			}
		}

		if options.Where != nil {
			for _, where := range options.Where {
				db.Where(where.Query, where.Args...)
			}
		}

		if options.Having != nil {
			for _, having := range options.Having {
				db.Having(having.Query, having.Args...)
			}
		}

		if options.Order != "" {
			db.Order(options.Order)
		}

		if options.Omit != "" {
			db.Omit(options.Omit)
		}

		if options.Limit != 0 {
			db.Limit(options.Limit)
		}

		if options.Offset != 0 {
			db.Offset(options.Offset)
		}

		return db
	}
}
