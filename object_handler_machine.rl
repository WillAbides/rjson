package rjson

func handleObjectValues(data []byte, handler ObjectValueHandler, stack []int) (int, []int ,error) {
  var top, cs, p, pp int
  var err error
  pe := len(data)
  eof := len(data)
  var currentFieldStart, currentFieldEnd int

%%{
machine handleObjectValues;

include skipper "skip_machine.rl";

prepush {
  if top + 1 >= len(stack) {
    stack = append(stack, make([]int, 1 + top - len(stack))...)
  }
}

action try_handler {
  pp, err = handler.HandleObjectValue(data[currentFieldStart+1:currentFieldEnd-1], data[p:])
  if err != nil {
    return p + pp, stack, err
  }
  if pp != 0 {
    fexec p + pp - 1;
  }
}

action try_handler_simple {
  _, err = handler.HandleObjectValue(data[currentFieldStart+1:currentFieldEnd-1], data[p:])
  if err != nil {
    return p, stack, err
  }
}

skip_array := skip_array_def;
skip_object := skip_object_def;

handled_value =
 json_bool >(try_handler_simple)
 | json_null >(try_handler_simple)
 | json_number >(try_handler_simple)
 | json_string >(try_handler)
 | '[' >(try_handler) @{fcall skip_array;}
 | '{' >(try_handler) @{fcall skip_object;}
;

json_object_field = (json_string >{currentFieldStart = p} %{currentFieldEnd = p} );

main :=
  ( json_space* (
  json_null |
  ('{'
      (
        json_space* json_object_field json_space* ':' json_space* handled_value
        (json_space* ',' json_space* json_object_field json_space* ':' json_space* handled_value )*
      )?
  json_space* '}'))) @err{
    return p, stack, errInvalidObject
  };

write data; write init;  write exec;
}%%

return p, stack, err
}
