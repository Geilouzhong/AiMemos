import copy from "copy-to-clipboard";
import { CheckIcon, CopyIcon } from "lucide-react";
import React, { useEffect, useState } from "react";
import { toast } from "react-hot-toast";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { userServiceClient } from "@/connect";
import useCurrentUser from "@/hooks/useCurrentUser";
import useLoading from "@/hooks/useLoading";
import { handleError } from "@/lib/error";
import { CreatePersonalAccessTokenResponse } from "@/types/proto/api/v1/user_service_pb";
import { useTranslate } from "@/utils/i18n";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (response: CreatePersonalAccessTokenResponse) => void;
}

interface State {
  description: string;
  expiration: number;
}

interface CreatedTokenState {
  token: string;
  description: string;
}

function CreateAccessTokenDialog({ open, onOpenChange, onSuccess }: Props) {
  const t = useTranslate();
  const currentUser = useCurrentUser();
  const [state, setState] = useState({
    description: "",
    expiration: 30, // Default: 30 days
  });
  const [createdToken, setCreatedToken] = useState<CreatedTokenState | undefined>(undefined);
  const [copied, setCopied] = useState(false);
  const requestState = useLoading(false);

  useEffect(() => {
    if (!open) {
      setState({ description: "", expiration: 30 });
      setCreatedToken(undefined);
      setCopied(false);
    }
  }, [open]);

  // Expiration options in days (0 = never expires)
  const expirationOptions = [
    {
      label: t("setting.access-token-section.create-dialog.duration-1m"),
      value: 30,
    },
    {
      label: "90 Days",
      value: 90,
    },
    {
      label: t("setting.access-token-section.create-dialog.duration-never"),
      value: 0,
    },
  ];

  const setPartialState = (partialState: Partial<State>) => {
    setState({
      ...state,
      ...partialState,
    });
  };

  const handleDescriptionInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPartialState({
      description: e.target.value,
    });
  };

  const handleRoleInputChange = (value: string) => {
    setPartialState({
      expiration: Number(value),
    });
  };

  const handleSaveBtnClick = async () => {
    if (!state.description) {
      toast.error(t("message.description-is-required"));
      return;
    }

    try {
      requestState.setLoading();
      const response = await userServiceClient.createPersonalAccessToken({
        parent: currentUser?.name,
        description: state.description,
        expiresInDays: state.expiration,
      });

      requestState.setFinish();
      onSuccess(response);
      if (!response.token) {
        toast.error("Token creation succeeded but token value is missing.");
        onOpenChange(false);
        return;
      }
      setCreatedToken({ token: response.token, description: state.description });
    } catch (error: unknown) {
      handleError(error, toast.error, {
        context: "Create access token",
        onError: () => requestState.setError(),
      });
    }
  };

  const handleCopyToken = () => {
    if (!createdToken?.token) {
      return;
    }
    if (copy(createdToken.token)) {
      setCopied(true);
      toast.success(t("setting.access-token-section.access-token-copied-to-clipboard"));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t("setting.access-token-section.create-dialog.create-access-token")}</DialogTitle>
        </DialogHeader>
        {createdToken ? (
          <div className="flex flex-col gap-4">
            <div className="rounded-md border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
              {t("setting.access-token-section.create-dialog.access-token-created", { description: createdToken.description })}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="created-token">Token</Label>
              <div className="flex gap-2">
                <Input id="created-token" type="text" value={createdToken.token} readOnly className="font-mono text-xs" />
                <Button type="button" variant="outline" onClick={handleCopyToken} className="shrink-0">
                  {copied ? <CheckIcon className="w-4 h-4" /> : <CopyIcon className="w-4 h-4" />}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">This token is only shown once. Copy it before closing this dialog.</p>
            </div>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="grid gap-2">
              <Label htmlFor="description">
                {t("setting.access-token-section.create-dialog.description")} <span className="text-destructive">*</span>
              </Label>
              <Input
                id="description"
                type="text"
                placeholder={t("setting.access-token-section.create-dialog.some-description")}
                value={state.description}
                onChange={handleDescriptionInputChange}
              />
            </div>
            <div className="grid gap-2">
              <Label>
                {t("setting.access-token-section.create-dialog.expiration")} <span className="text-destructive">*</span>
              </Label>
              <RadioGroup value={state.expiration.toString()} onValueChange={handleRoleInputChange} className="flex flex-row gap-4">
                {expirationOptions.map((option) => (
                  <div key={option.value} className="flex items-center space-x-2">
                    <RadioGroupItem value={option.value.toString()} id={`expiration-${option.value}`} />
                    <Label htmlFor={`expiration-${option.value}`}>{option.label}</Label>
                  </div>
                ))}
              </RadioGroup>
            </div>
          </div>
        )}
        <DialogFooter>
          <Button variant="ghost" disabled={requestState.isLoading} onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          {createdToken ? (
            <Button onClick={() => onOpenChange(false)}>{t("common.confirm")}</Button>
          ) : (
            <Button disabled={requestState.isLoading} onClick={handleSaveBtnClick}>
              {t("common.create")}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default CreateAccessTokenDialog;
