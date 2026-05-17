// dao/hero_query.go 英雄属性、英雄信息、技能相关配置表查询
package info

import (
	"context"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== hero_lv ====================
func GetAllHeroLv(ctx context.Context) ([]*table.HeroLv, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_lv")
	dest := []*table.HeroLv{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_attr ====================
func GetAllHeroAttr(ctx context.Context) ([]*table.HeroAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_attr")
	dest := []*table.HeroAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_stage_consume ====================
func GetAllHeroStageConsume(ctx context.Context) ([]*table.HeroStageConsume, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_stage_consume")
	dest := []*table.HeroStageConsume{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_star_up_attr ====================
func GetAllHeroStarUpAttr(ctx context.Context) ([]*table.HeroStarUpAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_star_up_attr")
	dest := []*table.HeroStarUpAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_star_basic_attr ====================
func GetAllHeroStarBasicAttr(ctx context.Context) ([]*table.HeroStarUpAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_star_basic_attr")
	dest := []*table.HeroStarUpAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_stage_up_attr ====================
func GetAllHeroStageUpAttr(ctx context.Context) ([]*table.HeroStageUpAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_stage_up_attr")
	dest := []*table.HeroStageUpAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_stage_basic_attr ====================
func GetAllHeroStageBasicAttr(ctx context.Context) ([]*table.HeroStageUpAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_stage_basic_attr")
	dest := []*table.HeroStageUpAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_lv_basic_attr ====================
func GetAllHeroLvBasicAttr(ctx context.Context) ([]*table.HeroLvBasicAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_lv_basic_attr")
	dest := []*table.HeroLvBasicAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_star_consume ====================
func GetAllHeroStarConsume(ctx context.Context) ([]*table.HeroStarConsume, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_star_consume")
	dest := []*table.HeroStarConsume{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_star ====================
func GetAllHeroStar(ctx context.Context) ([]*table.HeroStar, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_star")
	dest := []*table.HeroStar{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== hero_info ====================
func GetAllHeroInfo(ctx context.Context) ([]*table.HeroInfo, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_info")
	dest := []*table.HeroInfo{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== skill_info ====================
func GetAllSkillInfoOrderByLevelDesc(ctx context.Context) ([]*table.SkillInfo, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("skill_info").Order("level", true)
	dest := []*table.SkillInfo{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== skill_target_choose ====================
func GetSkillTargetChooseById(ctx context.Context, id int) (*table.SkillTargetChoose, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("skill_target_choose").AndEqual("id", id).Limit(1)
	dest := &table.SkillTargetChoose{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	if dest.Id == 0 {
		return nil, nil
	}
	return dest, nil
}

// ==================== skill_buff ====================
func GetSkillBuffById(ctx context.Context, id int) (*table.SkillBuff, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("skill_buff").AndEqual("id", id).Limit(1)
	dest := &table.SkillBuff{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	if dest.Id == 0 {
		return nil, nil
	}
	return dest, nil
}
