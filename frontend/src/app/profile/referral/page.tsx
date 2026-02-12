"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/lib/auth-context";
import {
  getReferralInfo,
  applyReferralCode,
  type ReferralInfo,
} from "@/lib/api";
import {
  ArrowLeft,
  Copy,
  Check,
  Share2,
  Gift,
  Users,
  Wallet,
  Loader2,
} from "lucide-react";

export default function ReferralPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();

  const [referral, setReferral] = useState<ReferralInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState(false);
  const [copiedLink, setCopiedLink] = useState(false);

  // Apply referral code form
  const [applyCode, setApplyCode] = useState("");
  const [applying, setApplying] = useState(false);
  const [applyError, setApplyError] = useState("");
  const [applySuccess, setApplySuccess] = useState(false);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.replace("/login?redirect=/profile/referral");
      return;
    }

    getReferralInfo()
      .then(setReferral)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [isAuthenticated, authLoading, router]);

  const handleCopyCode = () => {
    if (!referral) return;
    navigator.clipboard.writeText(referral.referralCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCopyLink = () => {
    if (!referral) return;
    navigator.clipboard.writeText(referral.referralLink);
    setCopiedLink(true);
    setTimeout(() => setCopiedLink(false), 2000);
  };

  const handleShare = async () => {
    if (!referral) return;
    if (navigator.share) {
      try {
        await navigator.share({
          title: "Авангард 3D — реферальная ссылка",
          text: `Зарегистрируйся по моей ссылке и получи бонус! Код: ${referral.referralCode}`,
          url: referral.referralLink,
        });
      } catch {
        handleCopyLink();
      }
    } else {
      handleCopyLink();
    }
  };

  const handleApply = async (e: React.FormEvent) => {
    e.preventDefault();
    setApplyError("");
    setApplySuccess(false);
    if (!applyCode.trim()) return;

    setApplying(true);
    try {
      await applyReferralCode(applyCode.trim());
      setApplySuccess(true);
      setApplyCode("");
      // Refresh referral info
      const info = await getReferralInfo();
      setReferral(info);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Не удалось применить код";
      setApplyError(msg);
    } finally {
      setApplying(false);
    }
  };

  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <Link
        href="/profile"
        className="mb-6 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Назад к профилю
      </Link>

      <h1 className="mb-6 text-2xl font-bold">Реферальная программа</h1>

      <div className="space-y-6">
        {/* Stats */}
        {referral && (
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="flex items-center gap-3 rounded-lg border p-4">
              <Wallet className="h-5 w-5 text-primary" />
              <div>
                <p className="text-xs text-muted-foreground">Баланс</p>
                <p className="text-lg font-bold">
                  {referral.bonusBalance.toFixed(0)} &#8381;
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-4">
              <Users className="h-5 w-5 text-primary" />
              <div>
                <p className="text-xs text-muted-foreground">Рефералов</p>
                <p className="text-lg font-bold">{referral.referralsCount}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-lg border p-4">
              <Gift className="h-5 w-5 text-primary" />
              <div>
                <p className="text-xs text-muted-foreground">Ваш код</p>
                <p className="text-lg font-bold font-mono">
                  {referral.referralCode}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Your referral code */}
        {referral && (
          <div className="rounded-lg border p-6 space-y-4">
            <h2 className="text-lg font-semibold">Пригласите друзей</h2>
            <p className="text-sm text-muted-foreground">
              Поделитесь кодом или ссылкой. Ваш друг получит приветственный
              бонус, а вы будете получать процент от суммы его заказов!
            </p>

            <div className="space-y-3">
              <div>
                <Label className="text-xs text-muted-foreground">Код</Label>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 rounded-lg bg-muted px-4 py-2.5 font-mono">
                    {referral.referralCode}
                  </code>
                  <Button variant="outline" size="sm" onClick={handleCopyCode}>
                    {copied ? (
                      <Check className="h-4 w-4" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <div>
                <Label className="text-xs text-muted-foreground">Ссылка</Label>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 truncate rounded-lg bg-muted px-4 py-2.5 text-xs">
                    {referral.referralLink}
                  </code>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleCopyLink}
                  >
                    {copiedLink ? (
                      <Check className="h-4 w-4" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <Button onClick={handleShare} className="w-full">
                <Share2 className="mr-2 h-4 w-4" />
                Поделиться
              </Button>
            </div>
          </div>
        )}

        {/* Apply someone else's code */}
        <div className="rounded-lg border p-6 space-y-4">
          <h2 className="text-lg font-semibold">Ввести реферальный код</h2>
          <p className="text-sm text-muted-foreground">
            Если вас пригласил друг, введите его код и получите приветственный
            бонус.
          </p>

          <form onSubmit={handleApply} className="space-y-3">
            {applyError && (
              <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                {applyError}
              </div>
            )}
            {applySuccess && (
              <div className="flex items-center gap-2 rounded-lg bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400">
                <Check className="h-4 w-4" />
                Код применён! Бонус начислен.
              </div>
            )}

            <div className="flex items-center gap-2">
              <Input
                placeholder="REF-XXXXXX"
                value={applyCode}
                onChange={(e) => setApplyCode(e.target.value.toUpperCase())}
                className="font-mono"
              />
              <Button type="submit" disabled={applying || !applyCode.trim()}>
                {applying ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  "Применить"
                )}
              </Button>
            </div>
          </form>
        </div>

        {/* Link to bonus history */}
        <Link
          href="/profile/bonuses"
          className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-muted"
        >
          <Wallet className="h-5 w-5" />
          <span className="font-medium">История бонусов</span>
        </Link>
      </div>
    </div>
  );
}
