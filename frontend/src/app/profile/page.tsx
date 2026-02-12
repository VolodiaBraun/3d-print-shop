"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/lib/auth-context";
import { getProfile, updateProfile, type ProfileData } from "@/lib/api";
import {
  User,
  Mail,
  Phone,
  Calendar,
  Package,
  MessageSquare,
  Loader2,
  Check,
  Gift,
  Copy,
  Wallet,
} from "lucide-react";
import { getReferralInfo, type ReferralInfo } from "@/lib/api";

export default function ProfilePage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading, logout } = useAuth();

  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [phone, setPhone] = useState("");
  const [referral, setReferral] = useState<ReferralInfo | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.replace("/login?redirect=/profile");
      return;
    }

    getProfile()
      .then((data) => {
        setProfile(data);
        setFirstName(data.firstName || "");
        setLastName(data.lastName || "");
        setPhone(data.phone || "");
      })
      .catch(() => setError("Не удалось загрузить профиль"))
      .finally(() => setLoading(false));

    getReferralInfo()
      .then(setReferral)
      .catch(() => {});
  }, [isAuthenticated, authLoading, router]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess(false);
    setSaving(true);
    try {
      const updated = await updateProfile({ firstName, lastName, phone });
      setProfile(updated);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch {
      setError("Не удалось сохранить");
    } finally {
      setSaving(false);
    }
  };

  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-12 text-center">
        <p className="text-muted-foreground">{error || "Профиль не найден"}</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold">Профиль</h1>

      <div className="space-y-6">
        {/* Info cards */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="flex items-center gap-3 rounded-lg border p-4">
            <Mail className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">Email</p>
              <p className="text-sm font-medium">{profile.email || "Не указан"}</p>
            </div>
          </div>
          <div className="flex items-center gap-3 rounded-lg border p-4">
            <User className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">Telegram</p>
              <p className="text-sm font-medium">
                {profile.telegramId
                  ? `@${profile.username || profile.telegramId}`
                  : "Не привязан"}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3 rounded-lg border p-4">
            <Calendar className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">Дата регистрации</p>
              <p className="text-sm font-medium">
                {new Date(profile.createdAt).toLocaleDateString("ru-RU")}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3 rounded-lg border p-4">
            <User className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-xs text-muted-foreground">Роль</p>
              <p className="text-sm font-medium">
                {profile.role === "admin" ? "Администратор" : "Покупатель"}
              </p>
            </div>
          </div>
        </div>

        {/* Edit form */}
        <form onSubmit={handleSave} className="space-y-4 rounded-lg border p-6">
          <h2 className="text-lg font-semibold">Редактировать</h2>

          {error && (
            <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          )}
          {success && (
            <div className="flex items-center gap-2 rounded-lg bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400">
              <Check className="h-4 w-4" />
              Сохранено
            </div>
          )}

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="firstName">Имя</Label>
              <Input
                id="firstName"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="lastName">Фамилия</Label>
              <Input
                id="lastName"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="phone">Телефон</Label>
            <div className="relative">
              <Phone className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="phone"
                type="tel"
                className="pl-10"
                placeholder="+7 (999) 123-45-67"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
              />
            </div>
          </div>

          <Button type="submit" disabled={saving}>
            {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Сохранить
          </Button>
        </form>

        {/* Referral section */}
        {referral && (
          <div className="rounded-lg border p-6 space-y-4">
            <div className="flex items-center gap-2">
              <Gift className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold">Реферальная программа</h2>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center gap-3 rounded-lg bg-muted/50 p-4">
                <Wallet className="h-5 w-5 text-primary" />
                <div>
                  <p className="text-xs text-muted-foreground">Бонусный баланс</p>
                  <p className="text-lg font-bold">{referral.bonusBalance.toFixed(0)} &#8381;</p>
                </div>
              </div>
              <div className="flex items-center gap-3 rounded-lg bg-muted/50 p-4">
                <User className="h-5 w-5 text-primary" />
                <div>
                  <p className="text-xs text-muted-foreground">Приглашено друзей</p>
                  <p className="text-lg font-bold">{referral.referralsCount}</p>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Ваш реферальный код:</p>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded-lg bg-muted px-4 py-2 font-mono text-sm">
                  {referral.referralCode}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    navigator.clipboard.writeText(referral.referralCode);
                    setCopied(true);
                    setTimeout(() => setCopied(false), 2000);
                  }}
                >
                  {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>

            <Link
              href="/profile/referral"
              className="inline-block text-sm font-medium text-primary underline-offset-4 hover:underline"
            >
              Подробнее о реферальной программе &rarr;
            </Link>
          </div>
        )}

        {/* Quick links */}
        <div className="grid gap-3 sm:grid-cols-2">
          <Link
            href="/orders"
            className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-muted"
          >
            <Package className="h-5 w-5" />
            <span className="font-medium">Мои заказы</span>
          </Link>
          <Link
            href="/profile/reviews"
            className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-muted"
          >
            <MessageSquare className="h-5 w-5" />
            <span className="font-medium">Мои отзывы</span>
          </Link>
          <Link
            href="/profile/bonuses"
            className="flex items-center gap-3 rounded-lg border p-4 transition-colors hover:bg-muted"
          >
            <Wallet className="h-5 w-5" />
            <span className="font-medium">История бонусов</span>
          </Link>
        </div>

        <Button
          variant="outline"
          className="w-full"
          onClick={() => {
            logout();
            router.push("/");
          }}
        >
          Выйти из аккаунта
        </Button>
      </div>
    </div>
  );
}
