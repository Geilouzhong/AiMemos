import { Input } from "@/components/ui/input";
import { useTranslate } from "@/utils/i18n";
import { useEditorContext } from "../state";

export const EditorTitle = () => {
  const t = useTranslate();
  const { state, actions, dispatch } = useEditorContext();

  return (
    <Input
      className="w-full border-none bg-transparent px-0 text-lg font-semibold shadow-none focus-visible:ring-0"
      type="text"
      placeholder={t("common.title")}
      value={state.title}
      onChange={(e) => dispatch(actions.updateTitle(e.target.value))}
    />
  );
};
