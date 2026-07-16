import unittest

from alignment import SpeakerTurn, TimedText, assign_speakers, dialogue_text


class AlignmentTest(unittest.TestCase):
    def test_assigns_by_largest_overlap_and_renames_by_first_appearance(self):
        transcript = [
            TimedText(0.0, 1.2, "Добрый день."),
            TimedText(1.1, 2.8, "Здравствуйте."),
            TimedText(2.9, 3.5, "Чем могу помочь?"),
        ]
        turns = [
            SpeakerTurn(0.0, 1.0, "SPEAKER_07"),
            SpeakerTurn(1.0, 2.9, "SPEAKER_02"),
            SpeakerTurn(2.9, 3.6, "SPEAKER_07"),
        ]
        segments = assign_speakers(transcript, turns)
        self.assertEqual(["Спикер 1", "Спикер 2", "Спикер 1"], [item["speaker"] for item in segments])
        self.assertEqual(
            "Спикер 1: Добрый день.\nСпикер 2: Здравствуйте.\nСпикер 1: Чем могу помочь?",
            dialogue_text(segments),
        )

    def test_merges_adjacent_lines_from_same_speaker(self):
        segments = [
            {"speaker": "Спикер 1", "text": "Один."},
            {"speaker": "Спикер 1", "text": "Два."},
        ]
        self.assertEqual("Спикер 1: Один. Два.", dialogue_text(segments))


if __name__ == "__main__":
    unittest.main()
