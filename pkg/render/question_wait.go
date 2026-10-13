package render

import "strconv"

func (r *Renderer) renderQuestionWait(timed, untimed int, total float64, avg, max *float64) {
	r.writeLine("timed_answers: " + strconv.Itoa(timed) + ", untimed_answers: " + strconv.Itoa(untimed))
	optionalMs := func(value *float64) string {
		if value == nil {
			return "n/a"
		}
		return strconv.FormatFloat(*value, 'f', 3, 64)
	}
	r.writeLine("total_answer_wait_ms: " + strconv.FormatFloat(total, 'f', 3, 64) +
		", avg_answer_wait_ms: " + optionalMs(avg) + ", max_answer_wait_ms: " + optionalMs(max))
	r.writeLine("wait includes tool/UI/transport overhead; sums per-question intervals, not unique wall-clock idle time")
}
