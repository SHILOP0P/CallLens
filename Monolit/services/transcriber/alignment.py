from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class TimedText:
    start: float
    end: float
    text: str


@dataclass(frozen=True)
class SpeakerTurn:
    start: float
    end: float
    speaker: str


def assign_speakers(
    transcript: list[TimedText], turns: list[SpeakerTurn]
) -> list[dict[str, object]]:
    """Assign each ASR segment to the speaker with the largest time overlap."""
    labels: dict[str, str] = {}
    result: list[dict[str, object]] = []

    for segment in transcript:
        if segment.end <= segment.start or not segment.text.strip():
            continue
        speaker = _best_speaker(segment, turns)
        if speaker not in labels:
            labels[speaker] = f"Спикер {len(labels) + 1}"
        result.append(
            {
                "speaker": labels[speaker],
                "start_seconds": round(segment.start, 3),
                "end_seconds": round(segment.end, 3),
                "text": segment.text.strip(),
            }
        )
    return result


def dialogue_text(segments: list[dict[str, object]]) -> str:
    lines: list[str] = []
    for segment in segments:
        speaker = str(segment["speaker"])
        text = str(segment["text"]).strip()
        if not text:
            continue
        if lines and lines[-1].startswith(f"{speaker}: "):
            lines[-1] = f"{lines[-1]} {text}"
        else:
            lines.append(f"{speaker}: {text}")
    return "\n".join(lines)


def _best_speaker(segment: TimedText, turns: list[SpeakerTurn]) -> str:
    best_speaker = "SPEAKER_UNKNOWN"
    best_overlap = 0.0
    midpoint = (segment.start + segment.end) / 2
    for turn in turns:
        overlap = max(0.0, min(segment.end, turn.end) - max(segment.start, turn.start))
        if overlap > best_overlap:
            best_speaker, best_overlap = turn.speaker, overlap
        elif best_overlap == 0.0 and turn.start <= midpoint <= turn.end:
            best_speaker = turn.speaker
    return best_speaker
