import type { Memo, MemoRelation_Memo } from "@/types/proto/api/v1/memo_service_pb";

const extractFirstMeaningfulLine = (content: string): string => {
  for (const rawLine of content.split("\n")) {
    const line = rawLine.trim().replace(/^#+\s*/, "");
    if (line) {
      return line;
    }
  }
  return "";
};

export const getMemoDisplayTitle = (memo: Pick<Memo, "title" | "content" | "snippet">): string => {
  if (memo.title.trim()) {
    return memo.title.trim();
  }

  const firstLine = extractFirstMeaningfulLine(memo.content);
  if (firstLine) {
    return firstLine;
  }

  if (memo.snippet.trim()) {
    return memo.snippet.trim();
  }

  return "Untitled";
};

export const getRelatedMemoDisplayTitle = (memo: Pick<MemoRelation_Memo, "snippet">): string => {
  return memo.snippet.trim() || "Untitled";
};
