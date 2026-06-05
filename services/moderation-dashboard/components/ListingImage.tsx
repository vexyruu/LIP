"use client";

import Image from "next/image";
import { useState } from "react";

type ListingImageProps = {
  images?: string[];
  title: string;
  className?: string;
  priority?: boolean;
  sizes?: string;
  compact?: boolean;
};

function fallbackLetter(title: string): string {
  return title.trim().charAt(0).toUpperCase() || "?";
}

export function ListingImage({
  images,
  title,
  className = "",
  priority = false,
  sizes = "(max-width: 768px) 100vw, 50vw",
  compact = false,
}: ListingImageProps) {
  const urls = (images ?? []).filter(Boolean);
  const [activeIndex, setActiveIndex] = useState(0);
  const [broken, setBroken] = useState(false);

  const src = urls[activeIndex];
  const showImage = Boolean(src) && !broken;

  if (!showImage) {
    return (
      <div
        className={`flex items-center justify-center border border-surface-variant bg-surface-container-lowest ${className}`}
        role="img"
        aria-label={`No photo available for ${title}`}
      >
        <span className="text-4xl font-bold text-surface-container-high">
          {fallbackLetter(title)}
        </span>
      </div>
    );
  }

  return (
    <div className={`relative overflow-hidden border border-surface-variant bg-surface-container-lowest ${className}`}>
      <Image
        src={src}
        alt={title}
        fill
        priority={priority}
        sizes={sizes}
        className="object-cover"
        onError={() => setBroken(true)}
      />
      {urls.length > 1 && !compact && (
        <div className="absolute bottom-0 left-0 right-0 flex gap-2 border-t border-surface-variant/80 bg-background/80 p-2 backdrop-blur-sm">
          {urls.map((url, index) => (
            <button
              key={`${url}-${index}`}
              type="button"
              aria-label={`Show image ${index + 1} of ${urls.length}`}
              aria-pressed={index === activeIndex}
              onClick={() => {
                setBroken(false);
                setActiveIndex(index);
              }}
              className={`relative h-12 w-12 shrink-0 overflow-hidden border-2 ${
                index === activeIndex
                  ? "border-primary-container"
                  : "border-outline opacity-80 hover:opacity-100"
              }`}
            >
              <Image
                src={url}
                alt=""
                fill
                sizes="48px"
                className="object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export function listingPrimaryImage(images?: string[]): string | undefined {
  return images?.find(Boolean);
}
