package models

import (
	"sort"

	"gorm.io/gorm"
)

type ProxyTag struct {
	ProxyUuid string `gorm:"column:proxy_uuid;uniqueIndex:idx_proxy_tag" json:"proxy_uuid"` // 代理唯一标识
	TagUuid   string `gorm:"column:tag_uuid;uniqueIndex:idx_proxy_tag" json:"tag_uuid"`     // 标签唯一标识
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`
}

func (table *ProxyTag) TableName() string {
	return "proxy_tag"
}

func LoadTagMap() (map[string]*TagBasic, error) {
	list := make([]*TagBasic, 0)
	if err := DB.Model(new(TagBasic)).Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	tagMap := make(map[string]*TagBasic, len(list))
	for _, item := range list {
		tagMap[item.Uuid] = item
	}
	return tagMap, nil
}

func LoadProxyTagListMap(proxyUuids []string) (map[string][]TagBasic, error) {
	rels := make([]*ProxyTag, 0)
	tx := DB.Model(new(ProxyTag))
	if len(proxyUuids) > 0 {
		tx = tx.Where("proxy_uuid IN ?", proxyUuids)
	}
	if err := tx.Find(&rels).Error; err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return map[string][]TagBasic{}, nil
	}

	tagMap, err := LoadTagMap()
	if err != nil {
		return nil, err
	}

	proxyTagMap := make(map[string][]TagBasic)
	for _, rel := range rels {
		tag, ok := tagMap[rel.TagUuid]
		if !ok {
			continue
		}
		proxyTagMap[rel.ProxyUuid] = append(proxyTagMap[rel.ProxyUuid], *tag)
	}
	for proxyUuid, tagList := range proxyTagMap {
		sort.Slice(tagList, func(i, j int) bool {
			return tagList[i].Name < tagList[j].Name
		})
		proxyTagMap[proxyUuid] = tagList
	}
	return proxyTagMap, nil
}

func ReplaceProxyTags(tx *gorm.DB, proxyUuid string, tagUuidList []string) error {
	if err := tx.Where("proxy_uuid = ?", proxyUuid).Delete(new(ProxyTag)).Error; err != nil {
		return err
	}
	if len(tagUuidList) == 0 {
		return nil
	}
	rels := make([]*ProxyTag, 0, len(tagUuidList))
	for _, tagUuid := range tagUuidList {
		rels = append(rels, &ProxyTag{
			ProxyUuid: proxyUuid,
			TagUuid:   tagUuid,
		})
	}
	return tx.Create(&rels).Error
}
