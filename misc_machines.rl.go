// Code generated by script/generate-ragel-file misc_machines.rl. DO NOT EDIT.

package rjson

func unescapeStringContent(data []byte, dst []byte) ([]byte, int, error) {
	cs, p := 0, 0
	pe := len(data)
	eof := len(data)
	var segStart int
	dst = growBytesSliceCapacity(dst, len(dst)+len(data))
	var unescapeUnicodeCharBytes int
	var ok bool

	const unescapeStringContent_start int = 6
	const unescapeStringContent_first_final int = 6
	const unescapeStringContent_error int = 0

	const unescapeStringContent_en_main int = 6

	{
		cs = unescapeStringContent_start
	}

	{
		if p == pe {
			goto _test_eof
		}
		switch cs {
		case 6:
			goto st_case_6
		case 0:
			goto st_case_0
		case 7:
			goto st_case_7
		case 1:
			goto st_case_1
		case 8:
			goto st_case_8
		case 9:
			goto st_case_9
		case 10:
			goto st_case_10
		case 11:
			goto st_case_11
		case 12:
			goto st_case_12
		case 13:
			goto st_case_13
		case 14:
			goto st_case_14
		case 15:
			goto st_case_15
		case 16:
			goto st_case_16
		case 2:
			goto st_case_2
		case 3:
			goto st_case_3
		case 4:
			goto st_case_4
		case 5:
			goto st_case_5
		case 17:
			goto st_case_17
		}
		goto st_out
	st_case_6:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr0:

		return nil, p, errInvalidString

		goto st0
	st_case_0:
	st0:
		cs = 0
		goto _out
	tr15:
		segStart = p
		goto st7
	tr18:
		dst = append(dst, data[segStart:p]...)
		segStart = p
		goto st7
	st7:
		if p++; p == pe {
			goto _test_eof7
		}
	st_case_7:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr19
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr18
	tr17:
		segStart = p
		goto st1
	tr19:
		dst = append(dst, data[segStart:p]...)
		segStart = p
		goto st1
	st1:
		if p++; p == pe {
			goto _test_eof1
		}
	st_case_1:
		switch data[p] {
		case 34:
			goto tr1
		case 39:
			goto tr2
		case 47:
			goto tr3
		case 92:
			goto tr4
		case 98:
			goto tr5
		case 102:
			goto tr6
		case 110:
			goto tr7
		case 114:
			goto tr8
		case 116:
			goto tr9
		case 117:
			goto st2
		}
		goto tr0
	tr1:
		dst = append(dst, '"')
		goto st8
	st8:
		if p++; p == pe {
			goto _test_eof8
		}
	st_case_8:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr2:
		dst = append(dst, '\'')
		goto st9
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr3:
		dst = append(dst, '/')
		goto st10
	st10:
		if p++; p == pe {
			goto _test_eof10
		}
	st_case_10:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr4:
		dst = append(dst, '\\')
		goto st11
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr5:
		dst = append(dst, '\b')
		goto st12
	st12:
		if p++; p == pe {
			goto _test_eof12
		}
	st_case_12:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr6:
		dst = append(dst, '\f')
		goto st13
	st13:
		if p++; p == pe {
			goto _test_eof13
		}
	st_case_13:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr7:
		dst = append(dst, '\n')
		goto st14
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr8:
		dst = append(dst, '\r')
		goto st15
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	tr9:
		dst = append(dst, '\t')
		goto st16
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st3
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st3
			}
		default:
			goto st3
		}
		goto tr0
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st4
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st4
			}
		default:
			goto st4
		}
		goto tr0
	st4:
		if p++; p == pe {
			goto _test_eof4
		}
	st_case_4:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st5
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st5
			}
		default:
			goto st5
		}
		goto tr0
	st5:
		if p++; p == pe {
			goto _test_eof5
		}
	st_case_5:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto tr14
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto tr14
			}
		default:
			goto tr14
		}
		goto tr0
	tr14:

		dst, unescapeUnicodeCharBytes, ok = unescapeUnicodeChar(data[segStart:], dst)
		if !ok {
			return nil, p, errUnexpectedByteInString(data[p])
		}
		if unescapeUnicodeCharBytes > 6 {
			p += unescapeUnicodeCharBytes - 6
		}

		goto st17
	st17:
		if p++; p == pe {
			goto _test_eof17
		}
	st_case_17:
		switch data[p] {
		case 34:
			goto st0
		case 92:
			goto tr17
		}
		if data[p] <= 31 {
			goto st0
		}
		goto tr15
	st_out:
	_test_eof7:
		cs = 7
		goto _test_eof
	_test_eof1:
		cs = 1
		goto _test_eof
	_test_eof8:
		cs = 8
		goto _test_eof
	_test_eof9:
		cs = 9
		goto _test_eof
	_test_eof10:
		cs = 10
		goto _test_eof
	_test_eof11:
		cs = 11
		goto _test_eof
	_test_eof12:
		cs = 12
		goto _test_eof
	_test_eof13:
		cs = 13
		goto _test_eof
	_test_eof14:
		cs = 14
		goto _test_eof
	_test_eof15:
		cs = 15
		goto _test_eof
	_test_eof16:
		cs = 16
		goto _test_eof
	_test_eof2:
		cs = 2
		goto _test_eof
	_test_eof3:
		cs = 3
		goto _test_eof
	_test_eof4:
		cs = 4
		goto _test_eof
	_test_eof5:
		cs = 5
		goto _test_eof
	_test_eof17:
		cs = 17
		goto _test_eof

	_test_eof:
		{
		}
		if p == eof {
			switch cs {
			case 7:
				dst = append(dst, data[segStart:p]...)
			case 1, 2, 3, 4, 5:

				return nil, p, errInvalidString

			}
		}

	_out:
		{
		}
	}

	return dst, p, nil
}

func appendRemainderOfString(data []byte, dst []byte) ([]byte, int, error) {
	cs, p := 0, 0
	pe := len(data)
	eof := len(data)
	var segStart int
	dst = growBytesSliceCapacity(dst, len(dst)+len(data))
	var unescapeUnicodeCharBytes int
	var ok bool

	const appendRemainderOfString_start int = 1
	const appendRemainderOfString_first_final int = 17
	const appendRemainderOfString_error int = 0

	const appendRemainderOfString_en_main int = 1

	{
		cs = appendRemainderOfString_start
	}

	{
		if p == pe {
			goto _test_eof
		}
		switch cs {
		case 1:
			goto st_case_1
		case 0:
			goto st_case_0
		case 2:
			goto st_case_2
		case 17:
			goto st_case_17
		case 3:
			goto st_case_3
		case 4:
			goto st_case_4
		case 5:
			goto st_case_5
		case 6:
			goto st_case_6
		case 7:
			goto st_case_7
		case 8:
			goto st_case_8
		case 9:
			goto st_case_9
		case 10:
			goto st_case_10
		case 11:
			goto st_case_11
		case 12:
			goto st_case_12
		case 13:
			goto st_case_13
		case 14:
			goto st_case_14
		case 15:
			goto st_case_15
		case 16:
			goto st_case_16
		}
		goto st_out
	st_case_1:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr0:

		return nil, p, errInvalidString

		goto st0
	st_case_0:
	st0:
		cs = 0
		goto _out
	tr1:
		segStart = p
		goto st2
	tr4:
		dst = append(dst, data[segStart:p]...)
		segStart = p
		goto st2
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
		switch data[p] {
		case 34:
			goto tr5
		case 92:
			goto tr6
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr4
	tr5:
		dst = append(dst, data[segStart:p]...)
		goto st17
	st17:
		if p++; p == pe {
			goto _test_eof17
		}
	st_case_17:
		goto st0
	tr3:
		segStart = p
		goto st3
	tr6:
		dst = append(dst, data[segStart:p]...)
		segStart = p
		goto st3
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
		switch data[p] {
		case 34:
			goto tr7
		case 47:
			goto tr8
		case 92:
			goto tr9
		case 98:
			goto tr10
		case 102:
			goto tr11
		case 110:
			goto tr12
		case 114:
			goto tr13
		case 116:
			goto tr14
		case 117:
			goto st12
		}
		goto tr0
	tr7:
		dst = append(dst, '"')
		goto st4
	st4:
		if p++; p == pe {
			goto _test_eof4
		}
	st_case_4:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr8:
		dst = append(dst, '/')
		goto st5
	st5:
		if p++; p == pe {
			goto _test_eof5
		}
	st_case_5:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr9:
		dst = append(dst, '\\')
		goto st6
	st6:
		if p++; p == pe {
			goto _test_eof6
		}
	st_case_6:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr10:
		dst = append(dst, '\b')
		goto st7
	st7:
		if p++; p == pe {
			goto _test_eof7
		}
	st_case_7:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr11:
		dst = append(dst, '\f')
		goto st8
	st8:
		if p++; p == pe {
			goto _test_eof8
		}
	st_case_8:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr12:
		dst = append(dst, '\n')
		goto st9
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr13:
		dst = append(dst, '\r')
		goto st10
	st10:
		if p++; p == pe {
			goto _test_eof10
		}
	st_case_10:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	tr14:
		dst = append(dst, '\t')
		goto st11
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	st12:
		if p++; p == pe {
			goto _test_eof12
		}
	st_case_12:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st13
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st13
			}
		default:
			goto st13
		}
		goto tr0
	st13:
		if p++; p == pe {
			goto _test_eof13
		}
	st_case_13:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st14
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st14
			}
		default:
			goto st14
		}
		goto tr0
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto st15
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto st15
			}
		default:
			goto st15
		}
		goto tr0
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
		switch {
		case data[p] < 65:
			if 48 <= data[p] && data[p] <= 57 {
				goto tr19
			}
		case data[p] > 70:
			if 97 <= data[p] && data[p] <= 102 {
				goto tr19
			}
		default:
			goto tr19
		}
		goto tr0
	tr19:

		dst, unescapeUnicodeCharBytes, ok = unescapeUnicodeChar(data[segStart:], dst)
		if !ok {
			return nil, p, errUnexpectedByteInString(data[p])
		}
		if unescapeUnicodeCharBytes > 6 {
			p += unescapeUnicodeCharBytes - 6
		}

		goto st16
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
		switch data[p] {
		case 34:
			goto st17
		case 92:
			goto tr3
		}
		if data[p] <= 31 {
			goto tr0
		}
		goto tr1
	st_out:
	_test_eof2:
		cs = 2
		goto _test_eof
	_test_eof17:
		cs = 17
		goto _test_eof
	_test_eof3:
		cs = 3
		goto _test_eof
	_test_eof4:
		cs = 4
		goto _test_eof
	_test_eof5:
		cs = 5
		goto _test_eof
	_test_eof6:
		cs = 6
		goto _test_eof
	_test_eof7:
		cs = 7
		goto _test_eof
	_test_eof8:
		cs = 8
		goto _test_eof
	_test_eof9:
		cs = 9
		goto _test_eof
	_test_eof10:
		cs = 10
		goto _test_eof
	_test_eof11:
		cs = 11
		goto _test_eof
	_test_eof12:
		cs = 12
		goto _test_eof
	_test_eof13:
		cs = 13
		goto _test_eof
	_test_eof14:
		cs = 14
		goto _test_eof
	_test_eof15:
		cs = 15
		goto _test_eof
	_test_eof16:
		cs = 16
		goto _test_eof

	_test_eof:
		{
		}
		if p == eof {
			switch cs {
			case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16:

				return nil, p, errInvalidString
			}
		}

	_out:
		{
		}
	}

	return dst, p, nil
}

func countWhitespace(data []byte) int {
	cs, p := 0, 0
	pe := len(data)

	const countWhitespace_start int = 1
	const countWhitespace_first_final int = 1
	const countWhitespace_error int = 0

	const countWhitespace_en_main int = 1

	{
		cs = countWhitespace_start
	}

	{
		if p == pe {
			goto _test_eof
		}
		switch cs {
		case 1:
			goto st_case_1
		case 0:
			goto st_case_0
		case 2:
			goto st_case_2
		}
		goto st_out
	st_case_1:
		switch data[p] {
		case 13:
			goto st2
		case 32:
			goto st2
		}
		if 9 <= data[p] && data[p] <= 10 {
			goto st2
		}
		goto st0
	st_case_0:
	st0:
		cs = 0
		goto _out
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
		switch data[p] {
		case 13:
			goto st2
		case 32:
			goto st2
		}
		if 9 <= data[p] && data[p] <= 10 {
			goto st2
		}
		goto st0
	st_out:
	_test_eof2:
		cs = 2
		goto _test_eof

	_test_eof:
		{
		}
	_out:
		{
		}
	}

	return p
}
