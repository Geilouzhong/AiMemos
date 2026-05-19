package v1

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	v1pb "github.com/Geilouzhong/AiMemos/proto/gen/api/v1"
)

const memoMetadataPrefix = "<!-- aimemos-meta: "

var memoMetadataPattern = regexp.MustCompile(`(?s)\n?<!-- aimemos-meta: (\{.*\}) -->\s*$`)

type memoFileMetadata struct {
	MemoName   string `json:"memo_name"`
	UpdatedAt  string `json:"updated_at"`
	Title      string `json:"title"`
	Visibility string `json:"visibility"`
}

func ensureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func sanitizeMemoFileName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "untitled"
	}
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
	)
	cleaned := strings.TrimSpace(replacer.Replace(trimmed))
	if cleaned == "" {
		return "untitled"
	}
	return cleaned
}

func buildMemoMetadataComment(meta memoFileMetadata) (string, error) {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s -->", memoMetadataPrefix, string(metaBytes)), nil
}

func extractMemoTitleCandidate(memo *v1pb.Memo) string {
	if strings.TrimSpace(memo.Title) != "" {
		return memo.Title
	}
	for _, rawLine := range strings.Split(memo.Content, "\n") {
		line := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rawLine), "#"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	if strings.TrimSpace(memo.Snippet) != "" {
		return memo.Snippet
	}
	return strings.TrimPrefix(memo.Name, "memos/")
}

func resolvePullPath(requestedPath string, memo *v1pb.Memo) string {
	if strings.TrimSpace(requestedPath) == "" {
		fileName := sanitizeMemoFileName(extractMemoTitleCandidate(memo)) + ".md"
		return fileName
	}
	info, err := os.Stat(requestedPath)
	if err == nil && info.IsDir() {
		fileName := sanitizeMemoFileName(extractMemoTitleCandidate(memo)) + ".md"
		return filepath.Join(requestedPath, fileName)
	}
	if strings.HasSuffix(requestedPath, string(os.PathSeparator)) {
		fileName := sanitizeMemoFileName(extractMemoTitleCandidate(memo)) + ".md"
		return filepath.Join(requestedPath, fileName)
	}
	return requestedPath
}

func writeMemoToLocalFile(path string, memo *v1pb.Memo) (string, error) {
	resolvedPath := resolvePullPath(path, memo)
	if err := ensureParentDir(resolvedPath); err != nil {
		return "", err
	}
	meta := memoFileMetadata{
		MemoName:   memo.Name,
		Title:      memo.Title,
		Visibility: memo.Visibility.String(),
	}
	if memo.UpdateTime != nil {
		meta.UpdatedAt = memo.UpdateTime.AsTime().UTC().Format(time.RFC3339)
	}
	metaComment, err := buildMemoMetadataComment(meta)
	if err != nil {
		return "", err
	}

	content := strings.TrimRight(memo.Content, "\n") + "\n\n" + metaComment + "\n"
	if err := os.WriteFile(resolvedPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return resolvedPath, nil
}

func readMemoFromLocalFile(path string) (string, *memoFileMetadata, error) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	content := string(contentBytes)
	matches := memoMetadataPattern.FindStringSubmatch(content)
	if len(matches) != 2 {
		return strings.TrimRight(content, "\n"), nil, nil
	}
	meta := &memoFileMetadata{}
	if err := json.Unmarshal([]byte(matches[1]), meta); err != nil {
		return "", nil, err
	}
	body := memoMetadataPattern.ReplaceAllString(content, "")
	body = strings.TrimRight(body, "\n")
	return body, meta, nil
}

func updateMemoMetadataInLocalFile(path string, memo *v1pb.Memo) error {
	body, _, err := readMemoFromLocalFile(path)
	if err != nil {
		return err
	}
	resolvedPath, err := writeMemoToLocalFile(path, &v1pb.Memo{
		Name:       memo.Name,
		Title:      memo.Title,
		Content:    body,
		Visibility: memo.Visibility,
		UpdateTime: memo.UpdateTime,
	})
	if err != nil {
		return err
	}
	if resolvedPath != path {
		return fmt.Errorf("unexpected path change while updating metadata: %s", resolvedPath)
	}
	return nil
}

func checkMemoPushConflict(meta *memoFileMetadata, remoteMemo *v1pb.Memo) (bool, string) {
	if meta == nil || meta.UpdatedAt == "" || remoteMemo.UpdateTime == nil {
		return false, ""
	}
	localUpdatedAt, err := time.Parse(time.RFC3339, meta.UpdatedAt)
	if err != nil {
		return false, ""
	}
	remoteUpdatedAt := remoteMemo.UpdateTime.AsTime().UTC()
	if remoteUpdatedAt.After(localUpdatedAt) {
		return true, fmt.Sprintf("remote memo has been updated after pull: local=%s remote=%s", localUpdatedAt.Format(time.RFC3339), remoteUpdatedAt.Format(time.RFC3339))
	}
	return false, ""
}
