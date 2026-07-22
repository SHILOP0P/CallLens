package catalog

import (
	"bufio"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
)

//go:embed catalog.md
var markdown string

var anchorPattern = regexp.MustCompile(`^<a id="([pb][0-9]+)"></a>$`)

type Industry struct{ Key, Perspective, Title, BasePrompt string }
type Topic struct {
	Key, IndustryKey, Title, Prompt string
	SortOrder                       int
}

func Parse() ([]Industry, []Topic, error) {
	var industries []Industry
	var topics []Topic
	perspective := ""
	pendingKey, currentKey := "", ""
	positions := map[string]int{}
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == `<a id="personal"></a>` {
			perspective = "personal"
			continue
		}
		if line == `<a id="business"></a>` {
			perspective = "business"
			continue
		}
		if match := anchorPattern.FindStringSubmatch(line); len(match) == 2 {
			pendingKey = match[1]
			continue
		}
		if pendingKey != "" && strings.HasPrefix(line, "### ") {
			currentKey = pendingKey
			pendingKey = ""
			industries = append(industries, Industry{Key: currentKey, Perspective: perspective, Title: strings.TrimSpace(strings.TrimPrefix(line, "### "))})
			continue
		}
		if !strings.HasPrefix(line, "| ") || currentKey == "" {
			continue
		}
		parts := strings.SplitN(strings.Trim(line, "|"), "|", 3)
		if len(parts) != 3 {
			continue
		}
		kind, title, prompt := strings.TrimSpace(parts[0]), strings.Trim(strings.TrimSpace(parts[1]), "`"), strings.TrimSpace(parts[2])
		if kind == "Отраслевой" && title == "Базовый промпт" {
			industries[len(industries)-1].BasePrompt = prompt
			continue
		}
		if kind != "Ключевой промпт" || title == "" || prompt == "" {
			continue
		}
		positions[currentKey]++
		topics = append(topics, Topic{Key: fmt.Sprintf("%s-%03d", currentKey, positions[currentKey]), IndustryKey: currentKey, Title: title, Prompt: prompt, SortOrder: positions[currentKey]})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if len(industries) == 0 || len(topics) == 0 {
		return nil, nil, fmt.Errorf("catalog markdown has no industries or topics")
	}
	return industries, topics, nil
}

func Seed(db *sql.DB) error {
	industries, topics, err := Parse()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM prompt_topics WHERE topic_key ~ '^[pb][0-9]+-[0-9]+$'`); err != nil { return err }
	if _, err = tx.Exec(`DELETE FROM prompt_industries WHERE industry_key ~ '^[pb][0-9]+$'`); err != nil { return err }
	for index, industry := range industries {
		if _, err = tx.Exec(`INSERT INTO prompt_industries(industry_key,perspective,title,sort_order,base_prompt) VALUES($1,$2,$3,$4,$5) ON CONFLICT(industry_key) DO UPDATE SET perspective=EXCLUDED.perspective,title=EXCLUDED.title,sort_order=EXCLUDED.sort_order,base_prompt=EXCLUDED.base_prompt`, industry.Key, industry.Perspective, industry.Title, index+1, industry.BasePrompt); err != nil {
			return err
		}
	}
	for _, topic := range topics {
		if _, err = tx.Exec(`INSERT INTO prompt_topics(topic_key,industry_key,title,prompt_module,sort_order,is_active) VALUES($1,$2,$3,$4,$5,true) ON CONFLICT(topic_key) DO UPDATE SET industry_key=EXCLUDED.industry_key,title=EXCLUDED.title,prompt_module=EXCLUDED.prompt_module,sort_order=EXCLUDED.sort_order,is_active=true`, topic.Key, topic.IndustryKey, topic.Title, topic.Prompt, topic.SortOrder); err != nil {
			return err
		}
		if _, err = tx.Exec(`INSERT INTO prompt_topic_aliases(topic_key,normalized_phrase,weight,is_negative) VALUES($1,$2,1,false) ON CONFLICT(topic_key,normalized_phrase) DO NOTHING`, topic.Key, strings.ToLower(topic.Title)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
