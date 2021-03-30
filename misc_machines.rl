package rjson

func unescapeStringContent(data []byte, dst []byte) ([]byte, int, error) {
  cs, p := 0, 0
  pe := len(data)
  eof := len(data)
  var segStart int
  dst = growBytesSliceCapacity(dst, len(dst) + len(data))
  var unescapeUnicodeCharBytes int
  var ok bool

%%{
 machine unescapeStringContent;

 include common "common.rl";

 main :=
 (
 (
   not_double_quote_or_escape >{segStart = p} %{dst = append(dst, data[segStart:p] ...)}

     | /\\"/ @{dst = append(dst, '"')}
     | /\\\\/ @{dst = append(dst, '\\')}
     | /\\\// @{dst = append(dst, '/')}
     | /\\'/ @{dst = append(dst, '\'')}
     | /\\b/ @{dst = append(dst, '\b')}
     | /\\f/ @{dst = append(dst, '\f')}
     | /\\n/ @{dst = append(dst, '\n')}
     | /\\r/ @{dst = append(dst, '\r')}
     | /\\t/ @{dst = append(dst, '\t')}
     | escaped_unicode >{ segStart = p } @{
       dst, unescapeUnicodeCharBytes, ok = unescapeUnicodeChar(data[segStart:], dst)
       if !ok {
         return nil, p, errUnexpectedByteInString(data[p])
       }
       if unescapeUnicodeCharBytes > 6 {
         p += unescapeUnicodeCharBytes - 6
       }
     }
 )*) @err{
   return nil, p, errInvalidString
 };
 write data; write init; write exec;
 }%%

   return dst, p,nil
}

func appendRemainderOfString(data []byte, dst []byte) ([]byte, int, error) {
  cs, p := 0, 0
  pe := len(data)
  eof := len(data)
  var segStart int
  dst = growBytesSliceCapacity(dst, len(dst) + len(data))
  var unescapeUnicodeCharBytes int
  var ok bool

 %%{
 machine appendRemainderOfString;

 include common "common.rl";

 main :=
 (
 (
   not_double_quote_or_escape >{segStart = p} %{dst = append(dst, data[segStart:p] ...)}

     | /\\"/ @{dst = append(dst, '"')}
     | /\\\\/ @{dst = append(dst, '\\')}
     | /\\\// @{dst = append(dst, '/')}
     | /\\b/ @{dst = append(dst, '\b')}
     | /\\f/ @{dst = append(dst, '\f')}
     | /\\n/ @{dst = append(dst, '\n')}
     | /\\r/ @{dst = append(dst, '\r')}
     | /\\t/ @{dst = append(dst, '\t')}
     | escaped_unicode >{ segStart = p } @{
       dst, unescapeUnicodeCharBytes, ok = unescapeUnicodeChar(data[segStart:], dst)
       if !ok {
         return nil, p, errUnexpectedByteInString(data[p])
       }
       if unescapeUnicodeCharBytes > 6 {
         p += unescapeUnicodeCharBytes - 6
       }
     }
 )*
 double_quote) @err{
   return nil, p, errInvalidString
 };
 write data; write init; write exec;
 }%%

   return dst, p,nil
}
